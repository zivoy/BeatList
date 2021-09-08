package main

import (
	"log"

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
			defer func(closer fyne.URIReadCloser) {
				err := closer.Close()
				if err != nil {
					log.Println(err)
				}
			}(closer)

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
		activePlaylist.CustomData.ReadOnly = b
		changes(true)
	})
	ui.AllowDuplicates = widget.NewCheck("", func(b bool) {
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

	// songlist
	ui.songInfo = initSongList()
}

func (u UI) refresh() {
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
	u.ArchiveURL.SetText(activePlaylist.CustomData.ArchiveURL)

	u.songInfo.SongDiffChecks.Hide()
	u.songInfo.SongDiffText.Show()
	u.songInfo.SongDiffDropDown.Hide()
	u.songInfo.Songs.Select(0)
	u.songInfo.Songs.Unselect(0)
}
