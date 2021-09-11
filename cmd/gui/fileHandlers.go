package main

import (
	"log"

	"github.com/zivoy/BeatList/pkg/playlist"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
)

var playlistFilter = storage.NewExtensionFileFilter([]string{".json", ".bplist"})

func saveMenu() {
	d := dialog.NewFileSave(func(closer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, window)
			log.Println(err)
			return
		}
		if closer == nil {
			return
		}
		defer closeFile(closer)

		err = activePlaylist.SavePretty(closer)
		if err != nil {
			dialog.ShowError(err, window)
			log.Println(err)
			return
		}
		changes(false)
	}, window)
	d.SetFileName(lastOpened.Name())
	uri, err := storage.ListerForURI(getBaseDir(lastOpened))
	if err == nil {
		d.SetLocation(uri)
	}
	d.SetFilter(playlistFilter)
	size := window.Canvas().Size()
	d.Resize(fyne.NewSize(size.Width*.8, size.Height*.8))
	d.Show()
}

//todo open recents
func openMenu() {
	confirmUnedited(func() {
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

			p, err := playlist.Load(closer)
			if err != nil {
				dialog.ShowError(err, window)
				log.Println(err)
				return
			}

			// load info on songs
			go func() {
				loadAll(p.Songs)
				changes(false)
			}()

			lastOpened = closer.URI()
			fyne.CurrentApp().Preferences().SetString("lastOpened", lastOpened.String())
			activePlaylist = p
			ui.refresh()
			changes(false)
		}, window)
		//d.SetFileName(lastOpened.Name())
		uri, err := storage.ListerForURI(getBaseDir(lastOpened))
		if err == nil {
			d.SetLocation(uri)
		}
		d.SetFilter(playlistFilter)
		size := window.Canvas().Size()
		d.Resize(fyne.NewSize(size.Width*.8, size.Height*.8))
		//todo add beat saber folder to favourites
		d.Show()
	}, window)
}
