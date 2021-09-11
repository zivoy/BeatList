//go:generate fyne bundle -o bundled.go Icon.png

package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"fyne.io/fyne/v2/driver/desktop"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

const VERSION = "0.1.4"
const imageSize = 512 // 512 by 512 pixels
var BeatSaverRe = regexp.MustCompile(`(?i)(?:beatsaver\.com/maps/|!bsr )?([[0-9A-F]+).*`)

var window fyne.Window

//todo second main menu screen with selector and stuff
// create a transition effect using the virtual canvas
// todo make mobile friendly ver of ui
func main() {
	a := app.NewWithID("BeatList") //com.zivoy.beatlist
	window = a.NewWindow("BeatList")
	a.Settings().SetTheme(theme.DarkTheme())
	window.SetIcon(resourceIconPng)
	var w, h float64
	w = 920
	h = 705
	s := a.Preferences().StringWithFallback("size", fmt.Sprintf("%f,%f", w, h))
	size := strings.Split(s, ",")
	width, err := strconv.ParseFloat(size[0], 32)
	if err != nil {
		width = w
	}
	height, err := strconv.ParseFloat(size[1], 32)
	if err != nil {
		height = h
	}
	window.Resize(fyne.NewSize(float32(width), float32(height)))

	defer closeFile(initLogging(a))

	initUpdater("zivoy", "BeatList")
	initSongListFuncs()
	initStorage(a)

	initGUI()

	loadLastSession()

	window.SetMainMenu(getMainMenu(a))

	window.SetContent(NewSetMinSize(container.NewBorder(nil, ui.LoadingBar, nil, nil,
		container.NewVScroll(MakePlaylistContainer())), 400, 400))

	if outdated := IsOutdated(VERSION); outdated.Outdated {
		updateDialog(outdated.Current, VERSION, window)
	}

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

	window.SetMaster()
	window.ShowAndRun()
	cleanup(a)
}
