package playlist

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"strings"

	"github.com/nfnt/resize"

	"image/jpeg"
	"image/png"
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

// --

// GetBase64Image returns the base64 image string
func (c Cover) GetBase64Image() string {
	return string(c)
}

func (c Cover) String() string {
	return c.GetBase64Image()
}

// GetImage returns an image.Image object
func (c Cover) GetImage() image.Image {
	if c == "" {
		return DefaultImage()
	}
	str := strings.SplitN(c.GetBase64Image(), ",", 2)
	decoded, err := base64.StdEncoding.DecodeString(str[1])
	if err != nil {
		return DefaultImage()
	}

	img, _, err := image.Decode(bytes.NewReader(decoded))
	if err != nil {
		return DefaultImage()
	}
	return img
}

// ImageToCover reads an image object and makes a Cover
func ImageToCover(img image.Image) (Cover, error) {
	var base64Encoding string
	buf := new(bytes.Buffer)
	err := png.Encode(buf, img)
	base64Encoding += "data:image/png;base64,"
	if err != nil {
		return "", err
	}

	base64Encoding += base64.StdEncoding.EncodeToString(buf.Bytes())

	return Cover(base64Encoding), nil
}

// ReaderToCover takes an io.Reader and returns a Cover
func ReaderToCover(reader io.Reader) (Cover, error) {
	img, mimeType, err := image.Decode(reader)
	if err != nil {
		return "", err
	}

	var base64Encoding string
	buf := new(bytes.Buffer)
	switch mimeType {
	case "image/jpeg", "jpeg", "jpg":
		err = jpeg.Encode(buf, img, nil)
		base64Encoding += "data:image/jpeg;base64,"
	case "image/png", "png":
		err = png.Encode(buf, img)
		base64Encoding += "data:image/png;base64,"
	}
	if err != nil {
		return "", err
	}

	base64Encoding += base64.StdEncoding.EncodeToString(buf.Bytes())

	return Cover(base64Encoding), nil
}

// Rescale changes the scale of an image
func (c *Cover) Rescale(size uint) {
	img := c.GetImage()
	img = resize.Resize(size, size, img, resize.Bilinear)
	s, err := ImageToCover(img)
	if err == nil {
		*c = s
	}
}

// DefaultImage returns a default image
func DefaultImage() image.Image {
	return image.Rect(0, 0, 1, 1) //todo
}
