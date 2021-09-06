package beatsaver

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

const apiRoot = "https://api.beatsaver.com"

var client = http.Client{
	Timeout: time.Second,
}

type Map struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Uploader    struct {
		Id        int    `json:"id"`
		Name      string `json:"name"`
		UniqueSet bool   `json:"uniqueSet"`
		Hash      string `json:"hash"`
		Avatar    string `json:"avatar"`
		Type      string `json:"type"`
	} `json:"uploader"`
	Metadata struct {
		Bpm             float32 `json:"bpm"`
		Duration        int     `json:"duration"`
		SongName        string  `json:"songName"`
		SongSubName     string  `json:"songSubName"`
		SongAuthorName  string  `json:"songAuthorName"`
		LevelAuthorName string  `json:"levelAuthorName"`
	} `json:"metadata"`
	Stats struct {
		Plays     int     `json:"plays"`
		Downloads int     `json:"downloads"`
		Upvotes   int     `json:"upvotes"`
		Downvotes int     `json:"downvotes"`
		Score     float64 `json:"score"`
	} `json:"stats"`
	Uploaded   time.Time `json:"uploaded"`
	Automapper bool      `json:"automapper"`
	Ranked     bool      `json:"ranked"`
	Qualified  bool      `json:"qualified"`
	Versions   []struct {
		Hash        string       `json:"hash"`
		Key         string       `json:"key"`
		State       string       `json:"state"`
		CreatedAt   time.Time    `json:"createdAt"`
		SageScore   int          `json:"sageScore"`
		Diffs       []MapVersion `json:"diffs"`
		DownloadURL string       `json:"downloadURL"`
		CoverURL    string       `json:"coverURL"`
		PreviewURL  string       `json:"previewURL"`
	} `json:"versions"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	LastPublishedAt time.Time `json:"lastPublishedAt"`
}

type MapVersion struct {
	Njs            float32 `json:"njs"`
	Offset         float64 `json:"offset"`
	Notes          int     `json:"notes"`
	Bombs          int     `json:"bombs"`
	Obstacles      int     `json:"obstacles"`
	Nps            float64 `json:"nps"`
	Length         float64 `json:"length"`
	Characteristic string  `json:"characteristic"`
	Difficulty     string  `json:"difficulty"`
	Events         int     `json:"events"`
	Chroma         bool    `json:"chroma"`
	Me             bool    `json:"me"`
	Ne             bool    `json:"ne"`
	Cinema         bool    `json:"cinema"`
	Seconds        float64 `json:"seconds"`
	ParitySummary  struct {
		Errors int `json:"errors"`
		Warns  int `json:"warns"`
		Resets int `json:"resets"`
	} `json:"paritySummary"`
	Stars float64 `json:"stars"`
}

func GetMap(hash string) (Map, error) {
	return getMap(apiRoot + "/maps/hash/" + hash)
}

func GetMapFromID(id string) (Map, error) {
	return getMap(apiRoot + "/maps/id/" + id)
}

func getMap(url string) (Map, error) { // todo save to tmp folder
	m := Map{}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return m, err
	}

	req.Header.Set("User-Agent", "BeatList")

	res, err := client.Do(req)
	if err != nil {
		return m, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return m, err
	}

	err = json.Unmarshal(body, &m)
	if err != nil {
		return m, err
	}
	return m, nil
}
