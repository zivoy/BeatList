package playlist

//https://github.com/rithik-b/PlaylistManager/wiki/

// Cover type contains functions for base64 covers
type Cover string

type Playlist struct {
	Title       string      `json:"playlistTitle"`
	Author      string      `json:"playlistAuthor,omitempty"`
	Description string      `json:"playlistDescription,omitempty"`
	Cover       Cover       `json:"image,omitempty"` //base 64 encoded string
	Songs       []*Song     `json:"songs"`
	CustomData  *CustomData `json:"CustomData,omitempty"`
}

type CustomData struct {
	AllowDuplicates bool   `json:"AllowDuplicates"`
	ArchiveURL      string `json:"customArchiveURL,omitempty"` //customArchiveURL
	SyncURL         string `json:"syncURL,omitempty"`
	ReadOnly        bool   `json:"ReadOnly,omitempty"`
}

type Difficulties struct {
	Characteristic string `json:"characteristic"` // Standard, OneSaber, NoArrows, 360Degree, 90Degree, Lawless
	Name           string `json:"name"`           // easy, normal, hard, expert, expertPlus
}

const (
	DifficultyEasy       = "easy"
	DifficultyNormal     = "normal"
	DifficultyHard       = "hard"
	DifficultyExpert     = "expert"
	DifficultyExpertPlus = "expertPlus"

	CharacteristicStandard  = "Standard"
	CharacteristicOneSaber  = "OneSaber"
	CharacteristicNoArrows  = "NoArrows"
	Characteristic360Degree = "360Degree"
	Characteristic90Degree  = "90Degree"
	CharacteristicLawless   = "Lawless"
)

type Song struct {
	Hash            string          `json:"hash"`    //REQUIRED FOR CUSTOM SONGS
	LevelID         string          `json:"levelid"` // REQUIRED FOR OST SONGS
	BeatSaverKey    string          `json:"key,omitempty"`
	SongName        string          `json:"songName,omitempty"`
	LevelAuthorName string          `json:"levelAuthorName,omitempty"`
	Difficulties    []*Difficulties `json:"difficulties,omitempty"`
}
