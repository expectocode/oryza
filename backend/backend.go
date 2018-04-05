package main

import (
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"
	"time"
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
	return &Backend{fileroot, db}
}

func (b Backend) create_tables() {
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
						   ID INT NOT NULL,
						   Mime TEXT NOT NULL,
						   Display TEXT NOT NULL,
						   Path TEXT NOT NULL,
						   Size INTEGER NOT NULL,
						   URI TEXT NOT NULL,
						   UploaderID INTEGER NOT NULL,
						   ExtraInfo TEXT NOT NULl,
						   UploadTime INTEGER NOT NULL,
						   PRIMARY KEY (ID)) WITHOUT ROWID`)
	if err != nil {
		log.Fatal("Could not create table file: ", err)
	}
}

func (b Backend) UploadFile(w http.ResponseWriter, r *http.Request) {
	//TODO
	things := map[string]string{"test": "data"}
	json.NewEncoder(w).Encode(things)
}

func (b Backend) DeleteFile(w http.ResponseWriter, r *http.Request) {
	//TODO
	things := map[string]string{"test": mux.Vars(r)["fileid"]}
	json.NewEncoder(w).Encode(things)
}

func (b Backend) RegisterUser(w http.ResponseWriter, r *http.Request) {
	//TODO
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

func main() {
	b := NewBackend("testdb.db")
	b.create_tables()

	router := mux.NewRouter()
	router.HandleFunc("/upload", b.UploadFile).Methods("POST")
	router.HandleFunc("/register", b.RegisterUser).Methods("POST")
	router.HandleFunc("/{fileid}", b.DeleteFile).Methods("DELETE")
	router.HandleFunc("/{fileid}", b.GetFile).Methods("GET")

	srv := &http.Server{
		Handler:      router,
		Addr:         ":8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
