package main

import (
	"fmt"
	"log"
	"net/url"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func getMainMenu(a fyne.App) *fyne.MainMenu {
	return fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("New", newBlank),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Open", openMenu), // open recent ?

			fyne.NewMenuItem("Save", saveMenu)),
		fyne.NewMenu("Help",
			fyne.NewMenuItem("Check for updates", func() {

				if outdated := IsOutdated(VERSION); outdated.Outdated {
					updateDialog(outdated.Current, VERSION, window)
				} else {
					dialog.ShowInformation("Up to date",
						fmt.Sprintf("You are on the latest version (v%s)", VERSION), window)
				}
			}),
			fyne.NewMenuItem("Report Bug", func() {
				bugURL, _ := url.Parse("https://github.com/zivoy/BeatList/issues/new")
				bugInfo := fmt.Sprintf("\n\n\n### Enviorment\n**BeatList version:** %s\n**OS:** %s (%s)", VERSION, runtime.GOOS, runtime.GOARCH)
				if logs := readLogs(); logs != "" {
					bugInfo = fmt.Sprintf("%s\n### Logs:\n```\n%s\n```", bugInfo, logs)
				}

				q := bugURL.Query()
				q.Set("body", bugInfo)
				bugURL.RawQuery = q.Encode()

				err := a.OpenURL(bugURL)
				if err != nil {
					log.Println(err)
				}
			}),
			fyne.NewMenuItem("About", func() {
				aboutW := a.NewWindow("About")
				cont := container.NewVBox()
				l := widget.NewLabel("BeatList a BeatSaber playlist creator\nVersion: v" + VERSION)
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
	)
}
