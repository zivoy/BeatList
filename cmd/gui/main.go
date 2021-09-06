package main

import (
	"BeatList/internal/beatsaver"
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"io"
	"log"
	"math"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"fyne.io/fyne/v2/driver/desktop"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nfnt/resize"
	"github.com/tcnksm/go-latest"
	"github.com/zivoy/BeatList/pkg/playlist"

	"image/jpeg"
	"image/png"
)

const VERSION = "0.0.1"
const imageSize = 256 // 256 by 256 pixels
var BeatSaverRe = regexp.MustCompile(`(?i)(?:(?:beatsaver\.com/maps/)|(?:!bsr ))?([[0-9A-F]+).*`)
var playlistFilter = storage.NewExtensionFileFilter([]string{".json", ".bplist"})

var activePlaylist playlist.Playlist

var lastOpened fyne.URI
var window fyne.Window
var defaultLoc fyne.URI
var CacheDir fyne.URI

type SongDiffs struct {
	chars []string
	diffs map[string][5]bool
}

var songDiffs = map[string]SongDiffs{}
var Diffs = []string{
	playlist.DifficultyEasy,
	playlist.DifficultyNormal,
	playlist.DifficultyHard,
	playlist.DifficultyExpert,
	playlist.DifficultyExpertPlus,
}

var githubTag = &latest.GithubTag{
	Owner:             "zivoy",
	Repository:        "BeatList",
	FixVersionStrFunc: latest.DeleteFrontV(),
}

type UI struct {
	Title, Author, Description, SyncURL, ArchiveURL *widget.Entry
	Image                                           *canvas.Image
	AllowDuplicates, ReadOnly                       *widget.Check
	SongName, SongMapper, SongDiffText              *widget.Label
	SongId                                          *widget.Hyperlink
	SongDiffs, SongDiffChecks                       *fyne.Container
	SongDiffDropDown                                *widget.Select
}

var ui UI

func main() {
	a := app.NewWithID("BeatList")
	window = a.NewWindow("BeatList")
	a.Settings().SetTheme(theme.DarkTheme())
	window.SetIcon(resourceIconPng)
	window.Resize(fyne.NewSize(750, 950))

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	logFile, _ := storage.Child(a.Storage().RootURI(), "latest.log")
	_ = storage.Delete(logFile)
	f, err := storage.Writer(logFile)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer func(f fyne.URIWriteCloser) {
		err := f.Close()
		if err != nil {
			log.Println(err)
		}
	}(f)
	multi := io.MultiWriter(f, os.Stdout)
	log.SetOutput(multi)

	CacheDir = storage.NewFileURI(filepath.Join(os.TempDir(), "BeatList"))
	err = os.MkdirAll(CacheDir.Path(), os.ModePerm)
	if err != nil {
		CacheDir, _ = storage.Child(a.Storage().RootURI(), "Cache")
		_ = os.MkdirAll(CacheDir.Path(), os.ModePerm)
	}

	executable, err := os.Executable()
	loc := "playlist.bplist"
	if err == nil {
		loc = path.Join(filepath.Dir(executable), loc)
	}
	defaultLoc = storage.NewFileURI(loc)
	lastOpened, _ = storage.ParseURI(a.Preferences().StringWithFallback("lastOpened", defaultLoc.String()))

	ui = UI{}
	ui.Author = widget.NewEntry()
	ui.Author.OnChanged = func(s string) {
		activePlaylist.Author = s
		changes(true)
	}
	ui.Author.SetPlaceHolder("(Optional)")
	ui.Title = widget.NewEntry()
	ui.Title.OnChanged = func(s string) {
		activePlaylist.Title = s
		changes(true)
	}
	ui.Description = widget.NewMultiLineEntry()
	ui.Description.Wrapping = fyne.TextWrapWord
	ui.Description.OnChanged = func(s string) {
		activePlaylist.Description = s
		changes(true)
	}
	ui.Description.SetPlaceHolder("(Optional)")
	ui.Image = canvas.NewImageFromImage(image.Rect(0, 0, 0, 0))
	ui.Image.SetMinSize(fyne.NewSize(128, 128))

	Songs := widget.NewList(func() int {
		return len(activePlaylist.Songs)
	}, func() fyne.CanvasObject {
		return widget.NewLabel("")
	}, func(id widget.ListItemID, object fyne.CanvasObject) {
		song := activePlaylist.Songs[id]
		object.(*widget.Label).SetText(fmt.Sprintf("%s [%s]", song.SongName, song.LevelAuthorName)) //todo make this look like the list in game (in terms of formatting)
	})

	ui.SongName = widget.NewLabel("")
	ui.SongName.Wrapping = fyne.TextWrapWord
	ui.SongMapper = widget.NewLabel("")
	ui.SongMapper.Wrapping = fyne.TextWrapWord

	ui.SongDiffText = widget.NewLabel("")
	ui.SongDiffText.Wrapping = fyne.TextWrapWord
	CharacteristicOptions := make([]string, 0)
	ui.SongDiffDropDown = widget.NewSelect(CharacteristicOptions, func(_ string) {})
	ui.SongDiffDropDown.Hide()
	ui.SongDiffChecks = container.NewHBox( // todo change this to allow wrapping
		widget.NewCheck("Easy", func(_ bool) {}),
		widget.NewCheck("Normal", func(_ bool) {}),
		widget.NewCheck("Hard", func(_ bool) {}),
		widget.NewCheck("Expert", func(_ bool) {}),
		widget.NewCheck("Expert Plus", func(_ bool) {}),
	)
	ui.SongDiffChecks.Hide()
	ui.SongDiffs = container.NewVBox(ui.SongDiffText, ui.SongDiffDropDown, ui.SongDiffChecks)

	ui.SongId = widget.NewHyperlink("", nil)
	songInfo := widget.NewForm(
		widget.NewFormItem("Map name", ui.SongName),
		widget.NewFormItem("Mapper", ui.SongMapper),
		widget.NewFormItem("Highlighted Diffs", ui.SongDiffs), //todo check things with char dropdown
		widget.NewFormItem("Beatsaver ID", ui.SongId),
	)
	var selected widget.ListItemID
	var selectedS bool
	Songs.OnUnselected = func(id widget.ListItemID) {
		selectedS = false
	}
	Songs.OnSelected = func(id widget.ListItemID) {
		selected = id
		selectedS = true
		song := activePlaylist.Songs[id]

		ui.SongDiffChecks.Hide()
		ui.SongDiffText.Show()
		ui.SongDiffDropDown.Hide()

		// update info
		updateSongInfo(song)
		Songs.Refresh()

		if details, ok := songDiffs[song.Hash]; ok {
			ui.SongDiffChecks.Show()
			ui.SongDiffText.Hide()
			ui.SongDiffDropDown.Show()

			CharacteristicOptions = details.chars
			ui.SongDiffDropDown.Options = CharacteristicOptions
			ui.SongDiffDropDown.Selected = CharacteristicOptions[0]

			songDetails := details
			selectedSong := song
			ui.SongDiffDropDown.OnChanged = func(s string) {
				diffs := songDetails.diffs[s]

				for i, o := range ui.SongDiffChecks.Objects {
					object := o.(*widget.Check)
					if diffs[i] {
						object.Show() //todo diff names
						object.Checked = false
					} else {
						object.Hide()
						continue
					}

					selectedDiff := Diffs[i]
					for _, d := range selectedSong.Difficulties {
						if strings.EqualFold(d.Characteristic, s) && strings.EqualFold(selectedDiff, d.Name) {
							object.Checked = true
							break
						}
					}
					object.OnChanged = func(b bool) {
						if b {
							selectedSong.Difficulties = append(selectedSong.Difficulties, &playlist.Difficulties{
								Characteristic: s,
								Name:           selectedDiff,
							})
						} else {
							for i, n := range selectedSong.Difficulties {
								if n.Name == selectedDiff && n.Characteristic == s {
									selectedSong.Difficulties[i] = selectedSong.Difficulties[len(selectedSong.Difficulties)-1]
									selectedSong.Difficulties = selectedSong.Difficulties[:len(selectedSong.Difficulties)-1]
									break
								}
							}
						}
					}
					object.Refresh()
				}
			}
			ui.SongDiffDropDown.Refresh()
			ui.SongDiffDropDown.OnChanged(CharacteristicOptions[0])
		}

		ui.SongName.SetText(song.SongName)
		ui.SongMapper.SetText(song.LevelAuthorName)
		ui.SongId.SetText(song.BeatSaverKey)
		err = ui.SongId.SetURLFromString(fmt.Sprintf("https://beatsaver.com/maps/%s", song.BeatSaverKey))
		if err != nil {
			fmt.Println(err) //todo
		}
		diffs := make([]string, len(song.Difficulties))
		for i, s := range song.Difficulties {
			diffs[i] = fmt.Sprintf("%s (%s)", s.Name, s.Characteristic)
		}
		if len(diffs) == 0 {
			ui.SongDiffText.SetText("None")
		} else {
			ui.SongDiffText.SetText(strings.Join(diffs, " | "))
		}
	}

	songListBar := widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			id := widget.NewEntry()
			id.SetPlaceHolder("Beatsaver id")
			beatsaverUrl, _ := url.Parse("https://beatsaver.com/")
			dialog.ShowForm("Beatsaver url / id", "Add", "cancel", widget.NewForm(
				widget.NewFormItem("Get link", widget.NewHyperlink("Beatsaver", beatsaverUrl)),
				widget.NewFormItem("id", id)).Items,
				func(b bool) {
					if b && BeatSaverRe.MatchString(id.Text) {
						subMatch := BeatSaverRe.FindStringSubmatch(id.Text)
						m, err := beatsaver.GetMapFromID(subMatch[1])
						if err != nil || m.Id == "" {
							dialog.ShowInformation("Error", fmt.Sprintf("Failed to use \"%s\" as a id", id.Text), window)
							return
						}
						version := m.Versions[0]                                            // not sure about this
						activePlaylist.Songs = append(activePlaylist.Songs, &playlist.Song{ //todo check if already in if disable doups
							Hash:            version.Hash,
							BeatSaverKey:    m.Id,
							SongName:        m.Metadata.SongName,
							LevelAuthorName: m.Metadata.LevelAuthorName,
						})
						songDiffs[version.Hash] = AddVersionChars(version.Diffs)
						changes(true)
						Songs.Refresh()
					}
				}, window)
		}),
		widget.NewToolbarAction(theme.CancelIcon(), func() {
			if selectedS {
				activePlaylist.Songs = append(activePlaylist.Songs[:selected], activePlaylist.Songs[selected+1:]...)
				Songs.Select(widget.ListItemID(math.Max(0, float64(selected-1))))
				Songs.Refresh()
			}
		}),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.MoveUpIcon(), func() {
			if selected > 0 {
				activePlaylist.Songs[selected-1], activePlaylist.Songs[selected] = activePlaylist.Songs[selected], activePlaylist.Songs[selected-1]
				Songs.Select(selected - 1)
				Songs.Refresh()
			}
		}),
		widget.NewToolbarAction(theme.MoveDownIcon(), func() {
			if selected < len(activePlaylist.Songs)-1 {
				activePlaylist.Songs[selected+1], activePlaylist.Songs[selected] = activePlaylist.Songs[selected], activePlaylist.Songs[selected+1]
				Songs.Select(selected + 1)
				Songs.Refresh()
			}
		}),
	)

	var songContainer *fyne.Container
	if fyne.CurrentDevice().IsMobile() {
		songContainer = container.NewMax(container.NewVSplit(container.NewBorder(nil, songListBar, nil, nil, Songs), songInfo))
	} else {
		Split := container.NewHSplit(Songs, container.NewHScroll(songInfo))
		Split.Offset = .4
		songContainer = container.NewBorder(nil, songListBar, nil, nil, Split)
	}

	ui.ReadOnly = widget.NewCheck("", func(b bool) {
		activePlaylist.CustomData.ReadOnly = b
		changes(true)
	})
	ui.AllowDuplicates = widget.NewCheck("", func(b bool) {
		activePlaylist.CustomData.AllowDuplicates = b
		changes(true)
	})
	ui.SyncURL = widget.NewEntry()
	ui.SyncURL.OnChanged = func(s string) {
		activePlaylist.CustomData.SyncURL = s
		changes(true)
	}
	ui.SyncURL.SetPlaceHolder("(Optional)")
	ui.ArchiveURL = widget.NewEntry()
	ui.ArchiveURL.OnChanged = func(s string) {
		activePlaylist.CustomData.ArchiveURL = s
		changes(true)
	}
	ui.ArchiveURL.SetPlaceHolder("(Optional)")

	form := container.NewVBox(
		widget.NewCard("Info", "", widget.NewForm(
			widget.NewFormItem("Image", container.NewHBox(ui.Image, container.NewVBox(widget.NewButton("Change", func() {
				d := dialog.NewFileOpen(func(closer fyne.URIReadCloser, err error) {
					if err != nil {
						dialog.ShowError(err, window)
						return
					}
					if closer == nil {
						return
					}
					defer closer.Close()

					img, encString, err := imageToBase64(closer)
					if err != nil {
						return //todo
					}
					activePlaylist.Cover = encString
					ui.Image.Image = canvas.NewImageFromImage(img).Image
					ui.Image.Refresh()
					changes(true)
				}, window)
				d.SetFilter(storage.NewMimeTypeFileFilter([]string{"image/png", "image/jpeg"})) //"image/gif"}))
				//d.SetFileName(defaultLoc.)
				d.Show()
			})))),
			widget.NewFormItem("Title", ui.Title),
			widget.NewFormItem("Author", ui.Author),
			widget.NewFormItem("Description", ui.Description),
		)),
		widget.NewCard("Metadata", "", widget.NewForm(
			widget.NewFormItem("Read only", ui.ReadOnly),
			widget.NewFormItem("Allow duplicates", ui.AllowDuplicates),
			widget.NewFormItem("Sync URL", ui.SyncURL),
			widget.NewFormItem("Archive URL", ui.ArchiveURL),
		)),
		widget.NewCard("Songs", "", songContainer),
	)

	loadLastSession()

	window.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("New", newFile),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Open", openMenu), // open recent ?

			fyne.NewMenuItem("Save", saveMenu)),
		fyne.NewMenu("Help",
			fyne.NewMenuItem("Check for updates", func() {
				res, _ := latest.Check(githubTag, VERSION)
				if res.Outdated {
					updateDialog(res.Current)
				} else {
					dialog.ShowInformation("Up to date", "You are on the latest version", window)
				}
			}),
			fyne.NewMenuItem("About", func() {
				aboutW := a.NewWindow("About")
				cont := container.NewVBox()
				l := widget.NewLabel("BeatList a BeatSaber playlist creator")
				l.Alignment = fyne.TextAlignCenter
				l.Wrapping = fyne.TextWrapWord
				cont.Add(l)
				github, _ := url.Parse("https://github.com/zivoy/BeatList/")
				h := widget.NewHyperlink("BeatList GitHub page", github)
				h.Alignment = fyne.TextAlignCenter
				cont.Add(h)

				l = widget.NewLabel("You can get the PlayList manager mod to use the playlists in game")
				l.Alignment = fyne.TextAlignCenter
				l.Wrapping = fyne.TextWrapWord
				cont.Add(l)
				github, _ = url.Parse("https://github.com/rithik-b/PlaylistManager")
				h = widget.NewHyperlink("PlaylistManager GitHub page", github)
				h.Alignment = fyne.TextAlignCenter
				cont.Add(h)
				b := widget.NewButton("Ok", func() {
					aboutW.Close()
				})
				cont.Add(b)
				aboutW.SetContent(cont)
				aboutW.Resize(fyne.NewSize(500, 150))
				aboutW.Show()
			})),
	))

	ctrlS := desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: desktop.ControlModifier}
	ctrlO := desktop.CustomShortcut{KeyName: fyne.KeyO, Modifier: desktop.ControlModifier}
	ctrlN := desktop.CustomShortcut{KeyName: fyne.KeyN, Modifier: desktop.ControlModifier}
	window.Canvas().AddShortcut(&ctrlN, func(shortcut fyne.Shortcut) {
		newFile()
	})
	window.Canvas().AddShortcut(&ctrlO, func(shortcut fyne.Shortcut) {
		openMenu()
	})
	window.Canvas().AddShortcut(&ctrlS, func(shortcut fyne.Shortcut) {
		saveMenu()
	})

	window.SetContent(form)

	res, err := latest.Check(githubTag, VERSION)
	if err == nil && res.Outdated {
		updateDialog(res.Current)
	}

	window.SetOnClosed(cleanup)
	window.SetMaster()
	window.ShowAndRun()
}

func cleanup() {
	if changes() {
		temp, _ := storage.Child(fyne.CurrentApp().Storage().RootURI(), lastOpened.Name())
		fyne.CurrentApp().Preferences().SetString("lastOpened", temp.String())
		w, _ := storage.Writer(temp)
		_ = activePlaylist.Save(w)
		_ = w.Close()
	}
}

func UpdateTitle() {
	title := "BeatList"
	if changes() {
		window.SetTitle("*" + title)
	} else {
		window.SetTitle(title)
	}
}

func confirmUnedited(callback func(), window fyne.Window) {
	if changes() {
		dialog.ShowConfirm("Unsaved Changes",
			"Do you want to proceed?\nDoing so will overwrite any changes that were not saved", func(b bool) {
				if b {
					callback()
				}
			}, window)
	} else {
		callback()
	}
}

func changes(val ...bool) bool {
	if len(val) != 0 {
		v := val[0]
		fyne.CurrentApp().Preferences().SetBool("changes", v)
		UpdateTitle()
		return v
	}
	return fyne.CurrentApp().Preferences().BoolWithFallback("changes", false)
}

func loadLastSession() {
	if fyne.CurrentApp().Preferences().BoolWithFallback("changes", false) {
		r, err := storage.Reader(lastOpened)
		if err != nil {
			activePlaylist = playlist.EmptyPlaylist()
			return
		}
		activePlaylist, err = playlist.Load(r)
		if err != nil {
			activePlaylist = playlist.EmptyPlaylist()
			return
		}
		changes(true)

		// load info on songs
		go loadAll(activePlaylist.Songs)

		_ = r.Close()
		_ = storage.Delete(lastOpened)
		fyne.CurrentApp().Preferences().SetString("lastOpened", "")
		fyne.CurrentApp().Preferences().RemoveValue("lastOpened")
		lastOpened = defaultLoc
		ui.refresh()
		return
	}
	activePlaylist = playlist.EmptyPlaylist()
}

func saveMenu() {
	d := dialog.NewFileSave(func(closer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		if closer == nil {
			return
		}
		defer closer.Close()

		err = activePlaylist.SavePretty(closer)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		changes(false)
	}, window)
	d.SetFileName(lastOpened.Name())
	d.SetFilter(playlistFilter)
	d.Show()
}

func openMenu() {
	confirmUnedited(func() {
		d := dialog.NewFileOpen(func(closer fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			if closer == nil {
				return
			}
			defer closer.Close()

			p, err := playlist.Load(closer)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}

			// load info on songs
			go loadAll(p.Songs)

			lastOpened = closer.URI()
			activePlaylist = p
			ui.refresh()
			changes(false)
		}, window)
		//d.SetFileName(lastOpened.Name())
		d.SetFilter(playlistFilter)
		d.Show()
	}, window)
}

func newFile() {
	confirmUnedited(func() {
		lastOpened = defaultLoc
		activePlaylist = playlist.EmptyPlaylist()
		ui.refresh()
		changes(false)
	}, window)
}

func updateDialog(latest string) {
	dialog.ShowCustomConfirm("Update", "Download", "Don't update", widget.NewLabel(
		fmt.Sprintf("v%s is not latest, you should upgrade to v%s", VERSION, latest)), func(b bool) {
		if b {
			updateUrl, _ := url.Parse("https://github.com/zivoy/BeatList/releases/latest")
			err := fyne.CurrentApp().OpenURL(updateUrl)
			if err != nil {
				dialog.ShowError(err, window)
			}
		}
	}, window)
}

func (u UI) refresh() {
	ui.Author.SetText(activePlaylist.Author)
	ui.Description.SetText(activePlaylist.Description)
	ui.Title.SetText(activePlaylist.Title)
	ui.Image.Image = canvas.NewImageFromImage(decodeBase64(activePlaylist.Cover)).Image
	ui.Image.Refresh()

	ui.ReadOnly.Checked = activePlaylist.CustomData.ReadOnly
	ui.ReadOnly.Refresh()
	ui.AllowDuplicates.Checked = activePlaylist.CustomData.AllowDuplicates
	ui.AllowDuplicates.Refresh()
	ui.SyncURL.SetText(activePlaylist.CustomData.SyncURL)
	ui.ArchiveURL.SetText(activePlaylist.CustomData.ArchiveURL)

	// todo refresh the song diff stuff
	//ui.SongDiffs.Refresh()
}

func imageToBase64(reader io.Reader) (image.Image, string, error) {
	img, mimeType, err := image.Decode(reader)
	if err != nil {
		return nil, "", err
	}

	img = resize.Resize(imageSize, imageSize, img, resize.Bilinear)

	var base64Encoding string
	buf := new(bytes.Buffer)
	switch mimeType {
	case "image/jpeg":
		err = jpeg.Encode(buf, img, nil)
		base64Encoding += "data:image/jpeg;base64,"
	case "image/png":
		err = png.Encode(buf, img)
		base64Encoding += "data:image/png;base64,"
	}
	if err != nil {
		return img, "", err
	}

	base64Encoding += base64.StdEncoding.EncodeToString(buf.Bytes())

	return img, base64Encoding, nil
}

func decodeBase64(s string) image.Image {
	if s == "" {
		return image.Rect(0, 0, 0, 0)
	}
	str := strings.SplitN(s, ",", 2)
	unbased, err := base64.StdEncoding.DecodeString(str[1])
	if err != nil {
		return image.Rect(0, 0, 0, 0)
	}

	img, _, err := image.Decode(bytes.NewReader(unbased))
	if err != nil {
		return image.Rect(0, 0, 0, 0)
	}
	return img
}

func AddVersionChars(mapData []beatsaver.MapVersion) SongDiffs {
	availableChars := make([]string, 0)
	diffs := map[string][5]bool{}
	for _, i := range mapData {
		//p:=sort.SearchStrings(availableVars, i.Characteristic)
		found := false
		var char string
		if strings.EqualFold(i.Characteristic, playlist.Characteristic360Degree) {
			char = playlist.Characteristic360Degree
		} else if strings.EqualFold(i.Characteristic, playlist.CharacteristicNoArrows) {
			char = playlist.CharacteristicNoArrows
		} else if strings.EqualFold(i.Characteristic, playlist.CharacteristicLawless) {
			char = playlist.CharacteristicLawless
		} else if strings.EqualFold(i.Characteristic, playlist.Characteristic90Degree) {
			char = playlist.Characteristic90Degree
		} else if strings.EqualFold(i.Characteristic, playlist.CharacteristicOneSaber) {
			char = playlist.CharacteristicOneSaber
		} else if strings.EqualFold(i.Characteristic, playlist.CharacteristicStandard) {
			char = playlist.CharacteristicStandard
		}

		for _, j := range availableChars {
			if j == char { //maybe do binary search?
				found = true
				break
			}
		}
		if !found {
			availableChars = append(availableChars, char)
			sort.Strings(availableChars)
			diffs[i.Characteristic] = [5]bool{}
		}
		a := diffs[i.Characteristic]
		for idx, v := range Diffs {
			if strings.EqualFold(v, i.Difficulty) {
				a[idx] = true
			}
		}
		diffs[i.Characteristic] = a
	}

	return SongDiffs{chars: availableChars, diffs: diffs}
}

// todo add progress bar to caching
func updateSongInfo(s *playlist.Song) {
	songInf, _ := storage.Child(CacheDir, s.Hash)
	var mapInfo beatsaver.Map
	exists, err := storage.Exists(songInf)
	if err == nil && exists {
		r, _ := storage.Reader(songInf)
		mapInfo, err = beatsaver.ReadMap(r)
		_ = r.Close()
		if mapInfo.Id == "" || err != nil {
			_ = storage.Delete(songInf)
			log.Printf("%s encountered an error or was blank, %e", s.Hash, err)
			return
		}
	} else {
		if err != nil {
			log.Println(err)
		}
		mapInfo, err = beatsaver.GetMap(s.Hash)
		if err == nil {
			w, e1 := storage.Writer(songInf)
			if w != nil {
				defer func(w fyne.URIWriteCloser) {
					err := w.Close()
					if err != nil {
						log.Println(err)
					}
				}(w)
			}
			if e1 != nil {
				log.Println(e1)
			} else {
				e1 = mapInfo.StoreMap(w)
				if e1 != nil {
					log.Println(e1)
				}
			}
		}
	}

	if err == nil {
		if s.SongName != mapInfo.Metadata.SongName {
			s.SongName = mapInfo.Metadata.SongName
			changes(true)
		}
		if s.BeatSaverKey != mapInfo.Id {
			s.BeatSaverKey = mapInfo.Id
			changes(true)
		}
		if s.LevelAuthorName != mapInfo.Metadata.LevelAuthorName {
			s.LevelAuthorName = mapInfo.Metadata.LevelAuthorName
			changes(true)
		}

		for _, i := range mapInfo.Versions {
			if strings.EqualFold(i.Hash, s.Hash) {
				songDiffs[s.Hash] = AddVersionChars(i.Diffs)
				break
			}
		}
	} else {
		log.Println(err)
	}
}

// todo group requests
func loadAll(songs []*playlist.Song) {
	for _, s := range songs {
		updateSongInfo(s)
	}
}
