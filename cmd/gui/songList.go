package main

import (
	//"github.com/zivoy/BeatList/internal/beatsaver"
	"BeatList/internal/beatsaver"

	"github.com/zivoy/BeatList/pkg/playlist"

	"fmt"
	"log"
	"math"
	"net/url"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type SongListUI struct {
	SongName, SongMapper, SongDiffText  *widget.Label
	SongId                              *widget.Hyperlink
	SongDiffs, SongDiffChecks, songMeta *fyne.Container
	SongDiffDropDown                    *widget.Select
	Songs                               *widget.List
}

var songListUI SongListUI

var SongListCache map[string]SongListItemCache

var itemSelected bool
var itemSelectedId widget.ListItemID

func initSongList() *SongListUI {
	songListUI = SongListUI{}

	SongListCache = map[string]SongListItemCache{}
	makeSongList()

	songListUI.SongName = widget.NewLabel("")
	songListUI.SongName.Wrapping = fyne.TextWrapWord
	songListUI.SongMapper = widget.NewLabel("")
	songListUI.SongMapper.Wrapping = fyne.TextWrapWord

	songListUI.SongDiffText = widget.NewLabel("")
	songListUI.SongDiffText.Wrapping = fyne.TextWrapWord
	CharacteristicOptions := make([]string, 0)
	songListUI.SongDiffDropDown = widget.NewSelect(CharacteristicOptions, func(_ string) {})
	songListUI.SongDiffDropDown.Hide()
	songListUI.SongDiffChecks = container.NewHBox( // todo change this to allow wrapping
		widget.NewCheck("Easy", func(_ bool) {}),
		widget.NewCheck("Normal", func(_ bool) {}),
		widget.NewCheck("Hard", func(_ bool) {}),
		widget.NewCheck("Expert", func(_ bool) {}),
		widget.NewCheck("Expert Plus", func(_ bool) {}),
	)
	songListUI.SongDiffChecks.Hide()
	songListUI.SongDiffs = container.NewVBox(songListUI.SongDiffText, songListUI.SongDiffDropDown, songListUI.SongDiffChecks)

	songListUI.SongId = widget.NewHyperlink("", nil)
	songListUI.songMeta = NewSetMinSize(widget.NewForm(
		widget.NewFormItem("Map name", songListUI.SongName),
		widget.NewFormItem("Mapper", songListUI.SongMapper),
		widget.NewFormItem("Highlighted Diffs", songListUI.SongDiffs),
		widget.NewFormItem("Beatsaver ID", songListUI.SongId),
	), 0, 300)
	songListUI.Songs.OnUnselected = func(id widget.ListItemID) {
		itemSelected = false
		songListUI.SongDiffChecks.Hide()
		songListUI.SongDiffText.Show()
		songListUI.SongDiffDropDown.Hide()

		songListUI.SongName.SetText("")
		songListUI.SongDiffText.SetText("")
		songListUI.SongMapper.SetText("")
		songListUI.SongId.SetText("")
	}
	songListUI.Songs.OnSelected = func(id widget.ListItemID) {
		itemSelectedId = id
		itemSelected = true
		song := activePlaylist.Songs[id]

		songListUI.SongDiffChecks.Hide()
		songListUI.SongDiffText.Show()
		songListUI.SongDiffDropDown.Hide()

		// update info
		updateSongInfo(song)
		songListUI.Songs.Refresh()

		if details, ok := songDiffs.Load(song.Hash); ok {
			songListUI.SongDiffChecks.Show()
			songListUI.SongDiffText.Hide()
			songListUI.SongDiffDropDown.Show()

			CharacteristicOptions = details.chars
			songListUI.SongDiffDropDown.Options = CharacteristicOptions
			songListUI.SongDiffDropDown.Selected = CharacteristicOptions[0]

			songDetails := details
			selectedSong := song
			songListUI.SongDiffDropDown.OnChanged = func(s string) {
				diffs := songDetails.diffs[s]

				for i, o := range songListUI.SongDiffChecks.Objects {
					object := o.(*widget.Check)
					if diffs[i] {
						object.Show() //todo custom diff names
						object.SetChecked(false)
					} else {
						object.Hide()
						continue
					}

					selectedDiff := Diffs[i]
					for _, d := range selectedSong.Difficulties {
						if strings.EqualFold(d.Characteristic, s) && strings.EqualFold(selectedDiff, d.Name) {
							object.SetChecked(true)
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
				}
			}
			songListUI.SongDiffDropDown.Refresh()
			songListUI.SongDiffDropDown.OnChanged(CharacteristicOptions[0])
		}

		songListUI.SongName.SetText(song.SongName)
		songListUI.SongMapper.SetText(song.LevelAuthorName)
		songListUI.SongId.SetText(song.BeatSaverKey)
		err := songListUI.SongId.SetURLFromString(fmt.Sprintf("https://beatsaver.com/maps/%s", song.BeatSaverKey))
		if err != nil {
			log.Println(err)
		}
		diffs := make([]string, len(song.Difficulties))
		for i, s := range song.Difficulties {
			diffs[i] = fmt.Sprintf("%s (%s)", s.Name, s.Characteristic)
		}
		if len(diffs) == 0 {
			songListUI.SongDiffText.SetText("None")
		} else {
			songListUI.SongDiffText.SetText(strings.Join(diffs, " | "))
		}
	}

	return &songListUI
}

func makeSongList() {
	songListUI.Songs = widget.NewList(func() int {
		return len(activePlaylist.Songs)
	}, func() fyne.CanvasObject {
		return NewSongItem("Song Name", "Sub name", "Song Author", "Mapper", "")
	}, func(id widget.ListItemID, object fyne.CanvasObject) {
		song := activePlaylist.Songs[id]
		cont := object.(*fyne.Container)

		var title, subtitle, author, mapper, cover string
		title = song.SongName
		mapper = song.LevelAuthorName
		if details, ok := songDiffs.Load(song.Hash); ok {
			subtitle = details.SubName
			author = details.SongAuthor
			cover = details.Cover
		}

		var Element *fyne.Container
		var newEl bool
		hashEl := hash(title + subtitle + author + mapper + cover)
		if songD, ok := SongListCache[song.Hash]; ok {
			if songD.hash == hashEl {
				Element = songD.song
			} else {
				newEl = true
			}
		} else {
			newEl = true
		}
		if newEl {
			Element = NewSongItem(title, subtitle, author, mapper, cover)
			SongListCache[song.Hash] = SongListItemCache{hash: hashEl, song: Element}
		}
		cont.Objects = Element.Objects
		cont.Refresh()
	})
}

func makeSongListBar() *widget.Toolbar {
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			id := widget.NewEntry()
			//todo make a search item instead of linking to the beatsaver website
			id.SetPlaceHolder("Beatsaver ID or URL")
			beatsaverUrl, _ := url.Parse("https://beatsaver.com/")
			dialog.ShowForm("Beatsaver URL / ID", "Add", "cancel", widget.NewForm(
				widget.NewFormItem("Get link", NewSetMinSize(widget.NewHyperlink("Beatsaver", beatsaverUrl), 350, 0)),
				widget.NewFormItem("ID", id)).Items,
				func(b bool) {
					if b && BeatSaverRe.MatchString(id.Text) {
						subMatch := BeatSaverRe.FindStringSubmatch(id.Text)
						m, err := beatsaver.GetMapFromID(subMatch[1])
						if err != nil || m.Id == "" {
							dialog.ShowInformation("Error", fmt.Sprintf("Failed to use \"%s\" as a id", id.Text), window)
							return
						}
						version := m.Versions[0] // not sure about this
						if !activePlaylist.CustomData.AllowDuplicates {
							for _, s := range activePlaylist.Songs {
								if s.Hash == version.Hash {
									return
								}
							}
						}
						activePlaylist.Songs = append(activePlaylist.Songs, &playlist.Song{
							Hash:            version.Hash,
							BeatSaverKey:    m.Id,
							SongName:        m.Metadata.SongName,
							LevelAuthorName: m.Metadata.LevelAuthorName,
						})
						meta := AddVersionChars(version.Diffs)
						meta.SongAuthor = m.Metadata.SongAuthorName
						meta.SubName = m.Metadata.SongSubName
						meta.Cover = version.CoverURL
						songDiffs.Store(version.Hash, meta)
						changes(true)
						songListUI.Songs.Refresh()
					}
				}, window)
		}),
		widget.NewToolbarAction(theme.CancelIcon(), func() {
			if itemSelected {
				activePlaylist.Songs = append(activePlaylist.Songs[:itemSelectedId], activePlaylist.Songs[itemSelectedId+1:]...)
				songListUI.Songs.Select(widget.ListItemID(math.Max(0, float64(itemSelectedId-1))))
				songListUI.Songs.Refresh()
			}
		}),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.MoveUpIcon(), func() {
			if itemSelectedId > 0 {
				activePlaylist.Songs[itemSelectedId-1], activePlaylist.Songs[itemSelectedId] =
					activePlaylist.Songs[itemSelectedId], activePlaylist.Songs[itemSelectedId-1]
				songListUI.Songs.Select(itemSelectedId - 1)
				songListUI.Songs.Refresh()
			}
		}),
		widget.NewToolbarAction(theme.MoveDownIcon(), func() {
			if itemSelectedId < len(activePlaylist.Songs)-1 {
				activePlaylist.Songs[itemSelectedId+1], activePlaylist.Songs[itemSelectedId] =
					activePlaylist.Songs[itemSelectedId], activePlaylist.Songs[itemSelectedId+1]
				songListUI.Songs.Select(itemSelectedId + 1)
				songListUI.Songs.Refresh()
			}
		}),
	)
}
