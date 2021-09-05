package playlist

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

// Save writes out Playlist to a writer, all on one line
func (p Playlist) Save(writer io.Writer) error {
	p.sanitise()
	bytes, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return write(bytes, writer)
}

// SavePretty writes out Playlist to a writer with indentation
func (p Playlist) SavePretty(writer io.Writer) error {
	p.sanitise()
	bytes, err := json.MarshalIndent(p, "", "    ")
	if err != nil {
		return err
	}
	return write(bytes, writer)
}

func write(bytes []byte, writer io.Writer) error {
	_, err := writer.Write(bytes)
	return err
}

func (p *Playlist) sanitise() {
	if p.Songs == nil {
		p.Songs = make([]*Song, 0)
	}
	for _, v := range p.Songs {
		if v.LevelID == "" {
			v.LevelID = fmt.Sprintf("custom_level_%s", v.Hash)
		}
	}
}

// Load reads a playlist from a reader
func Load(reader io.Reader) (Playlist, error) {
	dat, err := ioutil.ReadAll(reader)
	if err != nil {
		return Playlist{}, err
	}

	playlist := EmptyPlaylist()
	err = json.Unmarshal(dat, &playlist)
	if err != nil {
		return playlist, err
	}

	// handle stuff in other places
	other := map[string]interface{}{}
	err = json.Unmarshal(dat, &other)
	if err != nil {
		return playlist, err
	}
	for k, v := range other {
		switch k {
		case "AllowDuplicates":
			playlist.CustomData.AllowDuplicates = v.(bool)
		case "customArchiveURL":
			playlist.CustomData.ArchiveURL = v.(string)
		case "syncURL":
			playlist.CustomData.SyncURL = v.(string)
		case "ReadOnly":
			playlist.CustomData.ReadOnly = v.(bool)
		}
		delete(other, k)
	}

	return playlist, err
}

// EmptyPlaylist makes an empty playlist with the defaults
func EmptyPlaylist() Playlist {
	return Playlist{CustomData: &CustomData{AllowDuplicates: true}, Songs: make([]*Song, 0)}
}
