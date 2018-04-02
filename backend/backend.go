package backend

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"time"
)

type File struct {
	Name       string
	Mime       string
	Display    string
	Path       string
	Size       int
	Uri        string
	User       int
	UploadTime time.Time
}

type User struct {
	ID         int
	Token      string
	TelegramID int // This is 0 if there isn't one.
}

type Backend struct {
	FileRoot string
	DB       *sql.DB
}

func NewBackend(db_path string) *Backend {
	db, err := sql.Open("sqlite3", db_path)
	if err != nil {
		log.Fatal(err)
	}
	return &Backend{os.Getenv("ORYZA_ROOT"), db}
}

func (b Backend) db_firstrun() {
	//setup the DB
}

/*
class Backend:
	def __init__(self, db_name):
		self.db = sqlite3.connect(db_name)
		self.fileroot = os.environ['ORYZA_ROOT']

	def upload(self, filename, mimetype, display, uploader, uploadtime):
		pass
*/
