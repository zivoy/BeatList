package main

import (
	"errors"
	"log"
	"net/url"
	"strings"

	"fyne.io/fyne/v2/container"

	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"github.com/zivoy/BeatList/pkg/playlist"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

type UI struct {
	Title, Author, Description, SyncURL, ArchiveURL *widget.Entry
	Image                                           *canvas.Image
	AllowDuplicates, ReadOnly                       *widget.Check
	songInfo                                        *SongListUI
	LoadingBar                                      *widget.ProgressBar
	ImageChangeButton                               *widget.Button
}

var ui UI

func MakePlaylistContainer() *fyne.Container {
	songListBar := makeSongListBar()

	Split := container.NewVSplit(container.NewBorder(nil, songListBar, nil, nil, ui.songInfo.Songs),
		container.NewHScroll(ui.songInfo.songMeta))
	Split.Offset = .6
	vertical := NewSetMinSize(Split, 0, 350)

	Split = container.NewHSplit(ui.songInfo.Songs, container.NewHScroll(ui.songInfo.songMeta))
	Split.Offset = .45
	horizontal := container.NewBorder(nil, songListBar, nil, nil, Split)

	var songContainer *fyne.Container
	if fyne.CurrentDevice().IsMobile() {
		songContainer = container.NewMax(vertical)
	} else {
		songContainer = NewSizeSwitcher(600, horizontal, vertical)
	}

	return container.NewBorder(
		NewShrinkingColumns(
			widget.NewCard("Info", "", widget.NewForm(
				widget.NewFormItem("Image", container.NewHBox(ui.Image, NewCenter(ui.ImageChangeButton, posCenter, posLeading))),
				widget.NewFormItem("Title", ui.Title),
				widget.NewFormItem("Author", ui.Author),
				widget.NewFormItem("Description", ui.Description),
			)),
			widget.NewCard("Metadata", "", widget.NewForm(
				widget.NewFormItem("Read only", ui.ReadOnly),
				widget.NewFormItem("Allow duplicates", ui.AllowDuplicates),
				widget.NewFormItem("Sync URL", NewSetMinSize(ui.SyncURL, 240, 0)),
				widget.NewFormItem("Archive URL", ui.ArchiveURL),
			))), nil, nil, nil,
		widget.NewCard("Songs", "", songContainer),
	)
}

func initGUI() {
	ui = UI{}
	ui.LoadingBar = widget.NewProgressBar()
	ui.LoadingBar.Hide()

	// playlist info
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
	ui.Image = canvas.NewImageFromImage(playlist.DefaultImage())
	ui.Image.Hide()
	ui.Image.SetMinSize(fyne.NewSize(128, 128))
	ui.ImageChangeButton = widget.NewButton("Change", func() {
		d := dialog.NewFileOpen(func(closer fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, window)
				log.Println(err)
				return
			}
			if closer == nil {
				return
			}
			defer closeFile(closer)

			cover, err := playlist.ReaderToCover(closer)
			if err != nil {
				log.Println(err)
			}
			cover.Rescale(imageSize)

			activePlaylist.Cover = cover
			ui.Image.Image = canvas.NewImageFromImage(cover.GetImage()).Image
			ui.Image.Refresh()
			ui.Image.Show()
			changes(true)
		}, window)
		d.SetFilter(storage.NewMimeTypeFileFilter([]string{"image/png", "image/jpeg"})) //"image/gif"}))
		//d.SetFileName(defaultLoc.)
		d.Show()
	})

	// metadata
	ui.ReadOnly = widget.NewCheck("", func(b bool) {
		if b != activePlaylist.CustomData.ReadOnly {
			changes(true)
		}
		activePlaylist.CustomData.ReadOnly = b
	})
	ui.AllowDuplicates = widget.NewCheck("", func(b bool) {
		if b != activePlaylist.CustomData.AllowDuplicates {
			changes(true)
		}
		activePlaylist.CustomData.AllowDuplicates = b
		if !b {
			items := map[string]bool{}
			for _, s := range activePlaylist.Songs {
				items[s.Hash] = true
			}
			songs := make([]*playlist.Song, len(items))
			i := 0
			for _, s := range activePlaylist.Songs {
				if items[s.Hash] {
					songs[i] = s
					i++
					items[s.Hash] = false
				}
			}
			activePlaylist.Songs = songs
		}
	})
	ui.SyncURL = widget.NewEntry()
	ui.SyncURL.OnChanged = func(s string) {
		activePlaylist.CustomData.SyncURL = s
		changes(true)
	}
	ui.SyncURL.SetPlaceHolder("(Optional)")
	ui.SyncURL.Validator = verifyUrl
	ui.ArchiveURL = widget.NewEntry()
	ui.ArchiveURL.OnChanged = func(s string) {
		activePlaylist.CustomData.ArchiveURL = s
		changes(true)
	}
	ui.ArchiveURL.SetPlaceHolder("(Optional)")
	ui.ArchiveURL.Validator = verifyUrl

	// songlist
	ui.songInfo = initSongList()
}

func (u UI) refresh() {
	UpdateTitle()

	u.Author.SetText(activePlaylist.Author)
	u.Description.SetText(activePlaylist.Description)
	u.Title.SetText(activePlaylist.Title)
	if activePlaylist.Cover == "" {
		u.Image.Hide()
	} else {
		u.Image.Image = canvas.NewImageFromImage(activePlaylist.Cover.GetImage()).Image
		u.Image.Refresh()
		u.Image.Show()
	}

	u.ReadOnly.SetChecked(activePlaylist.CustomData.ReadOnly)
	u.AllowDuplicates.SetChecked(activePlaylist.CustomData.AllowDuplicates)
	u.SyncURL.SetText(activePlaylist.CustomData.SyncURL)
	if u.SyncURL.Text == "" {
		u.SyncURL.SetValidationError(nil)
	}
	u.ArchiveURL.SetText(activePlaylist.CustomData.ArchiveURL)
	if u.ArchiveURL.Text == "" {
		u.ArchiveURL.SetValidationError(nil)
	}
	u.songInfo.SongDiffChecks.Hide()
	u.songInfo.SongDiffText.Show()
	u.songInfo.SongDiffDropDown.Hide()
	u.songInfo.Songs.Select(0)
	u.songInfo.Songs.Unselect(0)

	u.ArchiveURL.SetValidationError(nil)
}

func verifyUrl(s string) error {
	if s == "" {
		return nil
	}
	u, err := url.ParseRequestURI(s)
	if err != nil {
		return errors.New("bad url")
	}
	if !(u.Scheme == "https" || u.Scheme == "http") {
		return errors.New("bad url scheme")
	}
	if u.Host == "" || !strings.ContainsRune(u.Host, '.') {
		return errors.New("missing url host")
	}
	return nil
}
