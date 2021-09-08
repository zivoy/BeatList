package main

import (
	"fmt"
	"log"
	"net/url"

	"github.com/tcnksm/go-latest"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var githubTag *latest.GithubTag

func initUpdater(owner, repo string) {
	githubTag = &latest.GithubTag{
		Owner:             owner,
		Repository:        repo,
		FixVersionStrFunc: latest.DeleteFrontV(),
	}
}

func updateDialog(latest, current string, window fyne.Window) {
	dialog.ShowCustomConfirm("Update", "Download", "Don't update", widget.NewLabel(
		fmt.Sprintf("v%s is not latest, you should upgrade to v%s", current, latest)), func(b bool) {
		if b {
			updateUrl, _ := url.Parse("https://github.com/zivoy/BeatList/releases/latest")
			err := fyne.CurrentApp().OpenURL(updateUrl)
			if err != nil {
				dialog.ShowError(err, window)
				log.Println(err)
			}
		}
	}, window)
}

func IsOutdated(version string) *latest.CheckResponse {
	res, err := latest.Check(githubTag, version)
	if err != nil {
		log.Println(err)
		return &latest.CheckResponse{
			Current:  "N/A",
			Outdated: false,
		}
	}
	return res
}
