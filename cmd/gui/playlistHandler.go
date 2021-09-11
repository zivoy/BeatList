package main

import (
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2/dialog"
	"github.com/zivoy/BeatList/pkg/playlist"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

var activePlaylist playlist.Playlist

func newBlank() {
	confirmUnedited(func() {
		lastOpened = defaultLoc
		activePlaylist = playlist.EmptyPlaylist()
		ui.refresh()
		changes(false)
	}, window)
}

func loadLastSession() {
	defer ui.refresh()
	if fyne.CurrentApp().Preferences().BoolWithFallback("changes", false) {
		r, err := storage.Reader(lastOpened)
		if err != nil {
			activePlaylist = playlist.EmptyPlaylist()
			changes(false)
			return
		}
		activePlaylist, err = playlist.Load(r)
		closeFile(r)
		if err != nil {
			log.Println(err)
			activePlaylist = playlist.EmptyPlaylist()
			changes(false)
			return
		}
		changes(true)

		// load info on songs
		go loadAll(activePlaylist.Songs)

		err = storage.Delete(lastOpened)
		if err != nil {
			log.Println(err)
		}
		fyne.CurrentApp().Preferences().SetString("lastOpened", "")
		fyne.CurrentApp().Preferences().RemoveValue("lastOpened")
		lastOpened = defaultLoc
		return
	}
	activePlaylist = playlist.EmptyPlaylist()
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

func cleanup(a fyne.App) {
	w, h := window.Canvas().Size().Components()
	if !window.FullScreen() { //todo detect maximised
		a.Preferences().SetString("size", fmt.Sprintf("%f,%f", w, h))
	}
	if changes() {
		<-time.After(200 * time.Millisecond) // todo remove with fyne 2.1
		temp, err := storage.Child(a.Storage().RootURI(), lastOpened.Name())
		if err != nil {
			log.Println(err)
			return
		}
		a.Preferences().SetString("lastOpened", temp.String())
		w, err := storage.Writer(temp)
		defer closeFile(w)
		if err != nil {
			log.Println(err)
			return
		}
		err = activePlaylist.Save(w)
		if err != nil {
			log.Println(err)
		}
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
