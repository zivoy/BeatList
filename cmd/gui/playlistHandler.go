package main

import (
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

func changes(val ...bool) bool {
	if len(val) != 0 {
		v := val[0]
		fyne.CurrentApp().Preferences().SetBool("changes", v)
		UpdateTitle()
		return v
	}
	return fyne.CurrentApp().Preferences().BoolWithFallback("changes", false)
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
