package main

import (
	"regexp"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const VERSION = "0.1.1"
const imageSize = 512 // 512 by 512 pixels
var BeatSaverRe = regexp.MustCompile(`(?i)(?:beatsaver\.com/maps/|!bsr )?([[0-9A-F]+).*`)

var window fyne.Window

func main() {
	a := app.NewWithID("BeatList")
	window = a.NewWindow("BeatList")
	a.Settings().SetTheme(theme.DarkTheme())
	window.SetIcon(resourceIconPng)
	window.Resize(fyne.NewSize(920, 705))

	initLogging(a)

	initUpdater("zivoy", "BeatList")
	initSongListFuncs()
	initStorage(a)

	initGUI()

	songListBar := makeSongListBar()

	var songContainer *fyne.Container
	if fyne.CurrentDevice().IsMobile() { //todo make a new layout that hides one once size is reached
		songContainer = container.NewMax(container.NewVSplit(
			container.NewBorder(nil, songListBar, nil, nil, ui.songInfo.Songs), ui.songInfo.songMeta))
	} else {
		Split := container.NewHSplit(ui.songInfo.Songs, container.NewHScroll(ui.songInfo.songMeta))
		Split.Offset = .45
		songContainer = container.NewBorder(nil, songListBar, nil, nil, Split)
	}

	form := container.NewVBox(
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
			))),
		widget.NewCard("Songs", "", songContainer),
	)

	loadLastSession()

	window.SetMainMenu(getMainMenu(a))

	// registering shortcuts
	ctrlS := desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: desktop.ControlModifier}
	ctrlO := desktop.CustomShortcut{KeyName: fyne.KeyO, Modifier: desktop.ControlModifier}
	ctrlN := desktop.CustomShortcut{KeyName: fyne.KeyN, Modifier: desktop.ControlModifier}
	window.Canvas().AddShortcut(&ctrlN, func(shortcut fyne.Shortcut) {
		newBlank()
	})
	window.Canvas().AddShortcut(&ctrlO, func(shortcut fyne.Shortcut) {
		openMenu()
	})
	window.Canvas().AddShortcut(&ctrlS, func(shortcut fyne.Shortcut) {
		saveMenu()
	})

	window.SetContent(NewSetMinSize(container.NewBorder(nil, ui.LoadingBar, nil, nil, container.NewVScroll(form)), 600, 400))

	if outdated := IsOutdated(VERSION); outdated.Outdated {
		updateDialog(outdated.Current, VERSION, window)
	}

	window.SetOnClosed(cleanup)
	window.SetMaster()
	window.ShowAndRun()
}
