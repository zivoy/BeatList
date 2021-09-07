module BeatList

go 1.16

require (
	fyne.io/fyne/v2 v2.0.4
	github.com/google/go-github v17.0.0+incompatible // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/hashicorp/go-version v1.3.0 // indirect
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/tcnksm/go-latest v0.0.0-20170313132115-e3007ae9052e
	github.com/zivoy/BeatList v0.1.0
)

replace (
	github.com/zivoy/BeatList => ./
)