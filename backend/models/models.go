package models

import (
	"database/sql"
	"encoding/json"
	"log"
	"io"
	"os"
	"time"
	"net/http"
	"fmt"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type File struct {
	Mime       string
	Display    string
	Path       string
	Size       int
	Uri        string
	User       int
	ExtraInfo  string
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
		log.Fatal("Could not open DB: ", err)
	}
	fileroot := os.Getenv("ORYZA_ROOT")
	log.Printf("Oryza file root: %s", fileroot)
	b := Backend{fileroot, db}
	b.createTables()
	return &b
}

func (b Backend) createTables() {
	//setup the DB
	_, err := b.DB.Exec(`CREATE TABLE IF NOT EXISTS User (
						   ID INT NOT NULL,
						   Name TEXT NOT NULl,
						   UploadToken TEXT NOT NULL,
						   PRIMARY KEY (ID)) WITHOUT ROWID`)
	if err != nil {
		log.Fatal("Could not create table user: ", err)
	}
	_, err = b.DB.Exec(`CREATE TABLE IF NOT EXISTS File (
						   ID INTEGER PRIMARY KEY AUTOINCREMENT,
						   MimeType TEXT NOT NULL,
						   Path TEXT NOT NULL,
						   Size INTEGER NOT NULL,
						   URI TEXT NOT NULL,
						   UploaderID INTEGER NOT NULL,
						   ExtraInfo TEXT NOT NULl,
						   UploadTime INTEGER NOT NULL)`)
	if err != nil {
		log.Fatal("Could not create table file: ", err)
	}
}

func fail(w http.ResponseWriter, reason string) {
	resp := map[string]string{"success": "false", "reason": reason}
	json.NewEncoder(w).Encode(resp)
	return
}

func (b Backend) UploadFile(w http.ResponseWriter, r *http.Request) {
	//TODO
	log.Printf("%s\n%s", r, r.Body)

	// We need to extract our keys from the payload
	// Token, upload content, mimetype, extra info

	token := r.FormValue("token")
	if token == "" {
		fail(w, "token cannot be empty")
		return
	}

	mimetype := r.FormValue("mimetype")
	if mimetype == "" {
		fail(w, "mimetype cannot be empty")
		return
	}

	extrainfo := r.FormValue("extrainfo")

	upfile, upfileheader, err := r.FormFile("uploadfile")
	if err == http.ErrMissingFile {
		fail(w, "Must supply an upload file")
		return
	} else if err != nil || upfile == nil || upfileheader  == nil{
		fail(w,fmt.Sprintf("Error with upload file: %s", err))
		return
	}

	// Now get the other fields for the DB
	// Display, Path, Size, Uri, Uploader

	log.Println("Success!", token, mimetype, extrainfo)
	f, err := os.Create("/tmp/thing")
	if err != nil {
		log.Printf("Error making disk file on upload %s", err)
		fail(w, "Error saving file")
	}
	defer f.Close()
	io.Copy(f, upfile)
}

func (b Backend) DeleteFile(w http.ResponseWriter, r *http.Request) {
	//TODO
	things := map[string]string{"test": mux.Vars(r)["fileid"]}
	json.NewEncoder(w).Encode(things)
}

func (b Backend) RegisterUser(w http.ResponseWriter, r *http.Request) {
	//TODO
	password, exists := os.LookupEnv("ORYZA_ROOT_PASSWORD")
	if !exists {
		log.Println("No oryza admin password set!!")
		fail(w, "No admin password set")
		return
	}
	// Handle request POST body
	log.Println(password)
	things := map[string]string{"test": mux.Vars(r)["fileid"]}
	json.NewEncoder(w).Encode(things)
}

func (b Backend) GetFile(w http.ResponseWriter, r *http.Request) {
	//TODO
	things := map[string]string{"test": mux.Vars(r)["fileid"]}
	json.NewEncoder(w).Encode(things)
}

/*
	def upload(self, filename, mimetype, display, uploader, uploadtime):
		pass
*/
