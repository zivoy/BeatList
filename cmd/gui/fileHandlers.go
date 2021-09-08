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
		defer func(closer fyne.URIWriteCloser) {
			err := closer.Close()
			if err != nil {
				log.Println(err)
			}
		}(closer)

		err = activePlaylist.SavePretty(closer)
		if err != nil {
			dialog.ShowError(err, window)
			log.Println(err)
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
			activePlaylist = p
			ui.refresh()
			changes(false)
		}, window)
		//d.SetFileName(lastOpened.Name())
		//uri, _ := storage.ListerForURI(lastOpened)
		//if err != nil {
		//	return
		//}
		//d.SetLocation()
		d.SetFilter(playlistFilter)
		d.Show()
	}, window)
}
