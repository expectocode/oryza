package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/expectocode/oryza/backend/urlgen"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

const DomainName = "http://up.unix.porn"

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
	if !strings.HasSuffix(fileroot, "/") {
		fileroot = fileroot + "/"
	}
	log.Printf("Oryza file root: %s", fileroot)
	b := Backend{fileroot, db}
	b.createTables()
	urlgen.Setup()
	return &b
}

func (b Backend) createTables() {
	//setup the DB
	_, err := b.DB.Exec(`CREATE TABLE IF NOT EXISTS User (
						   ID INTEGER PRIMARY KEY AUTOINCREMENT,
						   Name TEXT NOT NULl,
						   UploadToken TEXT NOT NULL)`)
	if err != nil {
		log.Fatal("Could not create table user: ", err)
	}
	_, err = b.DB.Exec(`CREATE TABLE IF NOT EXISTS File (
						   ID INTEGER PRIMARY KEY AUTOINCREMENT,
						   MimeType TEXT NOT NULL,
						   Path TEXT NOT NULL,
						   Size INTEGER NOT NULL,
						   ShortURI TEXT NOT NULL,
						   LongURI TEXT NOT NULL,
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
	log.Printf("%s\n%s", r, r.Body)

	// We need to extract our keys from the payload
	// Token, upload content, mimetype, extra info
	// We could save the original filename, but throw this away for privacy
	// No file size checks here - talk to people if they use lots.

	file := File{}
	file.UploadTime = time.Now()

	token := r.FormValue("token")
	if token == "" {
		log.Println("token", token)
		fail(w, "token cannot be empty")
		return
	}

	file.Mime = r.FormValue("mimetype")
	if file.Mime == "" {
		fail(w, "mimetype cannot be empty")
		return
	}

	file.ExtraInfo = r.FormValue("extrainfo")

	upfile, upfileheader, err := r.FormFile("uploadfile")
	if err == http.ErrMissingFile {
		fail(w, "Must supply an upload file")
		return
	} else if err != nil || upfile == nil || upfileheader == nil {
		fail(w, fmt.Sprintf("Error with upload file: %s", err))
		log.Println("Error with upload file! %s", err)
		return
	}

	// Now get the other fields for the DB
	// Path, Size, shorturi, longuri, Uploader
	var uploaderid int
	err = b.DB.QueryRow("select id from user where uploadtoken = ?", token).Scan(&uploaderid)
	if err != nil {
		if err == sql.ErrNoRows {
			fail(w, "Invalid upload token!")
			return
		} else if err != nil {
			fail(w, fmt.Sprintf("Report this error: %s", err))
			log.Println("Error getting uploader id", err)
			return
		}
	}
	file.Uploader = uploaderid

	// Generate shorturi
	non_duplicate := false
	var num_repeats int
	for !non_duplicate {
		num_repeats++
		if num_repeats > 200 {
			fail(w, "Could not generate a random shortname")
			log.Println("wtf, could not generate random shortname")
			return
		}
		file.ShortUri = urlgen.RandAlphanum(4) // eg C3dR

		// Check if there is already a file with this name
		var id int
		err = b.DB.QueryRow("select id from file where ShortURI = ?",
			file.ShortUri).Scan(&id)
		if err == sql.ErrNoRows {
			// Great!
			non_duplicate = true
		} else if err != nil {
			fail(w, fmt.Sprintf("Report this error: %s", err))
			log.Println("Error getting shortname duplicity", err)
			return
		}
	}
	file.ShortUri += urlgen.GetExtension(file.Mime)

	// Save path without the root before it
	file.Path = fmt.Sprintf("%d/%s%s", file.Uploader, file.ShortUri)

	// Generate longuri
	non_duplicate = false
	num_repeats = 0
	for !non_duplicate {
		num_repeats++
		if num_repeats > 200 {
			fail(w, "Could not generate a random longname")
			log.Println("wtf, could not generate random longname")
			return
		}
		file.LongUri = urlgen.GenLongUri() // eg InputDeterministicGBFeed

		// Check if there is already a file with this name
		var id int
		err = b.DB.QueryRow("select id from file where LongURI = ?",
			file.LongUri).Scan(&id)
		if err == sql.ErrNoRows {
			// Great!
			non_duplicate = true
		} else if err != nil {
			fail(w, fmt.Sprintf("Report this error: %s", err))
			log.Println("Error getting longname duplicity", err)
			return
		}
	}
	file.LongUri += urlgen.GetExtension(file.Mime)

	log.Println("File", file)

	fullpath := b.FileRoot + file.Path // Fileroot should always end in /
	f, err := os.Create(fullpath)
	if err != nil {
		log.Printf("Error making disk file on upload: %s", err)
		fail(w, fmt.Sprintf("Error saving file. Report this: %s", err))
		return
	}
	defer f.Close()
	io.Copy(f, upfile)

	// Get size
	info, err := os.Stat(fullpath)
	if err != nil {
		fail(w, "Could not stat the file")
		log.Printf("Could not stat file, did not delete.")
		return
	}
	file.Size = info.Size() // bytes

	log.Println("Finishing with", file)
	// Finally, save this file to the DB.
	_, err = b.DB.Exec(`INSERT INTO File(MimeType, Path, Size, ShortURI, LongURI,
										UploaderID, ExtraInfo, UploadTime)
										VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		file.Mime, file.Path, file.Size, file.ShortUri, file.LongUri,
		file.Uploader, file.ExtraInfo, file.UploadTime.Unix())
	if err != nil {
		fail(w, fmt.Sprintf("Report this error! Could not save file: %s", err))
		log.Printf("Could not save file to DB: %s", err)
		return
	}
	// Finally, success.
	resp := map[string]string{"success": "true",
		"url": fmt.Sprintf("%s/%s", DomainName, file.ShortUri)}
	json.NewEncoder(w).Encode(resp)
}

func (b Backend) DeleteFile(w http.ResponseWriter, r *http.Request) {
	//TODO
	things := map[string]string{"test": mux.Vars(r)["fileid"]}
	json.NewEncoder(w).Encode(things)
}

func (b Backend) RegisterUser(w http.ResponseWriter, r *http.Request) {
	//TODO
	// Insert ID, name, and token into the DB, and make their dir.
	password, exists := os.LookupEnv("ORYZA_ROOT_PASSWORD")
	if !exists {
		log.Println("No oryza admin password set!!")
		fail(w, "No admin password set")
		return
	}

	pass := r.FormValue("password")
	if pass == "" {
		fail(w, "Password may not be empty")
		return
	}
	log.Println("Password", password, pass)

	name := r.FormValue("name")
	if name == "" {
		fail(w, "name may not be empty")
		return
	}

	// Generate token
	var token string
	non_duplicate := false
	var num_repeats int
	for !non_duplicate {
		num_repeats++
		if num_repeats > 200 {
			fail(w, "Could not generate a random token")
			log.Println("wtf, could not generate random token")
			return
		}
		token = urlgen.RandAlphanum(16) // eg nfPBwjJM1pKmWxOd

		// Check if there is already a token with this value
		var id int
		err := b.DB.QueryRow("select id from user where UploadToken = ?",
			token).Scan(&id)
		if err == sql.ErrNoRows {
			// Great!
			non_duplicate = true
		} else if err != nil {
			fail(w, fmt.Sprintf("token duplicity error: %s", err))
			log.Println("Error getting token duplicity", err)
			return
		}
	}

	_, err := b.DB.Exec("INSERT INTO User(Name, UploadToken) VALUES (?, ?)",
		name, token)
	if err != nil {
		fail(w, fmt.Sprintf("Could not register user: %s", err))
		log.Printf("Could not register user: %s", err)
		return
	}

	var uid int
	err = b.DB.QueryRow("select id from user where uploadtoken = ?", token).Scan(&uid)
	if err != nil {
		if err == sql.ErrNoRows {
			fail(w, "wtf, invalid upload token we just made")
			return
		} else if err != nil {
			fail(w, fmt.Sprintf("Report this error: %s", err))
			log.Println("Error getting new user id", err)
			return
		}
	}

	err = os.Mkdir(fmt.Sprintf("%s/%d", b.FileRoot, uid), os.ModeDir | 0777)
	if err != nil {
		log.Println("Error making user dir for ID %d: %s", uid, err)
	}

	// Success!
	things := map[string]string{"success": "true",
	"userid": fmt.Sprintf("%d", uid), "token": token}
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
