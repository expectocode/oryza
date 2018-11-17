package models

import (
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type File struct {
	ID         int
	Mime       string
	Path       string
	Size       int64
	ShortUri   string
	LongUri    string
	Uploader   int
	ExtraInfo  string
	UploadTime time.Time
}

type FileListing struct {
	URL          string `json:"url"`
	Size         int    `json:"size"`         // bytes
	DateUploaded int    `json:"dateuploaded"` // unix timestamp
	ShortURI     string `json:"shorturi"`
	LongURI      string `json:"longuri"`
	MimeType     string `json:"mimetype"`
	ExtraInfo    string `json:"extrainfo"`
	Deleted      bool   `json:"deleted"`
}
