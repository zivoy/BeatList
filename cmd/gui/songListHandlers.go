package main

import (
	"github.com/zivoy/BeatList/internal/beatsaver"
	"image/color"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/zivoy/BeatList/pkg/playlist"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

var images = map[string]*canvas.Image{}
var requestLimiterTimer <-chan time.Time
var songDiffs *SongDiffMap
var imageDownloader *sync.Cond

var Diffs = []string{
	playlist.DifficultyEasy,
	playlist.DifficultyNormal,
	playlist.DifficultyHard,
	playlist.DifficultyExpert,
	playlist.DifficultyExpertPlus,
}

type SongDiffs struct {
	chars      []string
	diffs      map[string][5]bool
	SubName    string
	SongAuthor string
	Cover      string //url
}

type SongListItemCache struct {
	hash string
	song *fyne.Container
}

type SongDiffMap struct {
	sync.RWMutex
	internal map[string]SongDiffs
}

func NewSongDiffMap() *SongDiffMap {
	return &SongDiffMap{
		internal: make(map[string]SongDiffs),
	}
}

func (rm *SongDiffMap) Load(key string) (value SongDiffs, ok bool) {
	rm.RLock()
	result, ok := rm.internal[strings.ToLower(key)]
	rm.RUnlock()
	return result, ok
}

func (rm *SongDiffMap) Delete(key string) {
	rm.Lock()
	delete(rm.internal, strings.ToLower(key))
	rm.Unlock()
}

func (rm *SongDiffMap) Store(key string, value SongDiffs) {
	rm.Lock()
	rm.internal[strings.ToLower(key)] = value
	rm.Unlock()
}

// --

func initSongListFuncs() {
	imageDownloader = sync.NewCond(&sync.RWMutex{})
	requestLimiterTimer = time.Tick(100 * time.Millisecond)
	songDiffs = NewSongDiffMap()
}

func NewSongItem(SongName, SongSubName, Author, Mapper, cover string) *fyne.Container {
	l := container.NewWithoutLayout()
	var width1, width2, height float32

	text := canvas.NewText(SongName, color.White)
	text.TextStyle.Bold = true
	l.Add(text)
	titleSize := text.MinSize()
	width1 += titleSize.Width
	height += titleSize.Height

	text = canvas.NewText(" "+SongSubName, color.White)
	text.Move(fyne.NewPos(titleSize.Width, 0))
	l.Add(text)
	width1 += text.MinSize().Width

	if Author != "" && Author != Mapper {
		text = canvas.NewText(Author+" [", color.Gray{Y: 0xb3})
		text.Move(fyne.NewPos(0, titleSize.Height))
		l.Add(text)
		musicianSize := text.MinSize()
		width2 += musicianSize.Width
		height += musicianSize.Height

		text = canvas.NewText(Mapper, color.RGBA{R: 0xa7, G: 0xd9, B: 0x34, A: 0xff})
		text.Move(fyne.NewPos(musicianSize.Width, titleSize.Height))
		l.Add(text)
		mapperSize := text.MinSize()
		width2 += mapperSize.Width

		text = canvas.NewText("]", color.Gray{Y: 0xb3})
		text.Move(fyne.NewPos(musicianSize.Width+mapperSize.Width, titleSize.Height))
		l.Add(text)
		width2 += text.MinSize().Width
	} else {
		text = canvas.NewText(Mapper, color.RGBA{R: 0xa7, G: 0xd9, B: 0x34, A: 0xff})
		text.Move(fyne.NewPos(0, titleSize.Height))
		l.Add(text)
		mapperSize := text.MinSize()
		width2 += mapperSize.Width
		height += mapperSize.Height
	}

	var img *canvas.Image
	if im, ok := images[cover]; ok { // dont make another image cover
		img = im
	} else {
		img = canvas.NewImageFromImage(playlist.DefaultImage())
		GetImage(cover, img)
		img.SetMinSize(fyne.NewSize(64, 64))
		images[cover] = img
	}

	return container.NewHBox(img,
		NewCenter(container.NewPadded(NewSetMinSize(l, fyne.Max(width1, width2), height)), posCenter, posLeading))

}

func GetImage(url string, img *canvas.Image) {
	if url == "" {
		return
	}

	go func() {
		c := getCached("img-"+hash(url),
			func(reader io.Reader) (interface{}, error) {
				a, err := ioutil.ReadAll(reader)
				return playlist.Cover(a), err
			},
			func(item interface{}) bool {
				return item.(playlist.Cover).String() == ""
			},
			func() (interface{}, error) {
				cover := fetchCover(url, func() bool {
					return !img.Visible()
				})
				if cover != "" {
					//resize to save space
					cover.Rescale(64)
				}
				return cover, nil
			},
			func(writer io.Writer, item interface{}) error {
				_, err := writer.Write([]byte(item.(playlist.Cover).GetBase64Image()))
				return err
			})

		if c != nil {
			cover := c.(playlist.Cover)
			img.Image = cover.GetImage()
			img.Refresh()
		}
	}()
}

var waiting bool

func fetchCover(url string, stop func() bool) playlist.Cover {
	imageDownloader.L.Lock()
	if waiting {
		imageDownloader.Wait()
	}
	if stop() {
		return ""
	}
	waiting = true

	defer func() {
		waiting = false
		imageDownloader.Signal()
		imageDownloader.L.Unlock()
	}()

	<-requestLimiterTimer

	response, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return ""
	}
	defer closeFile(response.Body)

	if response.StatusCode != 200 {
		log.Println(url, response.StatusCode)
		return ""
	}
	cover, err := playlist.ReaderToCover(response.Body)
	if err != nil {
		log.Println(err)
	}
	return cover
}

func updateSongInfo(s *playlist.Song) {
	m := getCached(s.Hash,
		func(reader io.Reader) (interface{}, error) {
			return beatsaver.ReadMap(reader)
		},
		func(item interface{}) bool {
			return item.(beatsaver.Map).Id == ""
		},
		func() (interface{}, error) {
			return beatsaver.GetMap(s.Hash)
		},
		func(writer io.Writer, item interface{}) error {
			return item.(beatsaver.Map).StoreMap(writer)
		})

	if m != nil {
		mapInfo := m.(beatsaver.Map)
		c := false
		if s.SongName != mapInfo.Metadata.SongName {
			s.SongName = mapInfo.Metadata.SongName
			changes(true)
			c = true
		}
		if s.BeatSaverKey != mapInfo.Id {
			s.BeatSaverKey = mapInfo.Id
			changes(true)
			c = true
		}
		if s.LevelAuthorName != mapInfo.Metadata.LevelAuthorName {
			s.LevelAuthorName = mapInfo.Metadata.LevelAuthorName
			changes(true)
			c = true
		}

		if _, ok := songDiffs.Load(s.Hash); !ok || c {
			for _, i := range mapInfo.Versions {
				if strings.EqualFold(i.Hash, s.Hash) {
					meta := AddVersionChars(i.Diffs)
					meta.SongAuthor = mapInfo.Metadata.SongAuthorName
					meta.SubName = mapInfo.Metadata.SongSubName
					meta.Cover = i.CoverURL
					songDiffs.Store(i.Hash, meta)
					break
				}
			}
		}
	} else {
		log.Println("error getting hash", s.Hash)
	}
}

func AddVersionChars(mapData []beatsaver.MapVersion) SongDiffs {
	availableChars := make([]string, 0)
	diffs := map[string][5]bool{}
	for _, i := range mapData {
		//p:=sort.SearchStrings(availableVars, i.Characteristic)
		found := false
		var char string
		if strings.EqualFold(i.Characteristic, playlist.Characteristic360Degree) {
			char = playlist.Characteristic360Degree
		} else if strings.EqualFold(i.Characteristic, playlist.CharacteristicNoArrows) {
			char = playlist.CharacteristicNoArrows
		} else if strings.EqualFold(i.Characteristic, playlist.CharacteristicLawless) {
			char = playlist.CharacteristicLawless
		} else if strings.EqualFold(i.Characteristic, playlist.Characteristic90Degree) {
			char = playlist.Characteristic90Degree
		} else if strings.EqualFold(i.Characteristic, playlist.CharacteristicOneSaber) {
			char = playlist.CharacteristicOneSaber
		} else if strings.EqualFold(i.Characteristic, playlist.CharacteristicStandard) {
			char = playlist.CharacteristicStandard
		} else if strings.EqualFold(i.Characteristic, playlist.CharacteristicLightshow) {
			char = playlist.CharacteristicLightshow
		} else {
			log.Printf("potentially unsupported characteristic \"%s\"", i.Characteristic)
			char = i.Characteristic
		}

		for _, j := range availableChars {
			if j == char { //maybe do binary search?
				found = true
				break
			}
		}
		if !found {
			availableChars = append(availableChars, char)
			sort.Strings(availableChars)
			diffs[i.Characteristic] = [5]bool{}
		}
		a := diffs[i.Characteristic]
		for idx, v := range Diffs {
			if strings.EqualFold(v, i.Difficulty) {
				a[idx] = true
			}
		}
		diffs[i.Characteristic] = a
	}

	return SongDiffs{chars: availableChars, diffs: diffs}
}

// todo group requests and lock to allow only one at a time
func loadAll(songs []*playlist.Song) {
	ui.LoadingBar.Show()
	ui.LoadingBar.Min = 0
	ui.LoadingBar.Max = float64(len(songs)) - 1
	for i, s := range songs {
		ui.LoadingBar.SetValue(float64(i))
		updateSongInfo(s)
	}
	ui.LoadingBar.Hide()
}
