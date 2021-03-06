package methods

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

	m "github.com/expectocode/oryza/backend/models"
	"github.com/expectocode/oryza/backend/safemime"
	"github.com/expectocode/oryza/backend/urlgen"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type Backend struct {
	FileRoot   string
	DomainName string
	DB         *sql.DB
}

func Setup(db_path string) *Backend {
	db, err := sql.Open("sqlite3", db_path)
	if err != nil {
		log.Fatal("Could not open DB: ", err)
	}
	fileroot := os.Getenv("ORYZA_ROOT")
	if fileroot == "" {
		log.Fatal("You must set $ORYZA_ROOT")
	}
	if !strings.HasSuffix(fileroot, "/") {
		fileroot = fileroot + "/"
	}
	DomainName := os.Getenv("ORYZA_DOMAIN_NAME")
	if DomainName == "" {
		log.Fatal("You must set $ORYZA_DOMAIN_NAME")
	}
	log.Printf("Oryza file root: %s", fileroot)
	b := Backend{fileroot, DomainName, db}
	b.createTables()
	urlgen.Setup()
	return &b
}

func fail(w http.ResponseWriter, reason string) {
	resp := map[string]string{"success": "false", "reason": reason}
	json.NewEncoder(w).Encode(resp)
	return
}

type ListingResponse struct {
	Uploads []m.FileListing `json:"uploads"`
	Success bool            `json:"success"`
}

func (b Backend) ListUploads(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	if token == "" {
		fail(w, "token cannot be empty")
		return
	}

	var uploaderID int
	err := b.DB.QueryRow("select id from user where uploadtoken = ?",
		token).Scan(&uploaderID)
	if err != nil {
		if err == sql.ErrNoRows {
			fail(w, "Invalid upload token!")
			return
		} else if err != nil {
			fail(w, fmt.Sprintf("user id db error: %s", err))
			log.Println("Error getting uploader id", err)
			return
		}
	}

	// Get info on all uploads from this user
	var listings []m.FileListing

	rows, err := b.DB.Query("select size, UploadTime, ShortURI, LongURI, MimeType, ExtraInfo, Deleted from File where uploaderid = ?", uploaderID)
	if err != nil {
		log.Println("Error retrieving rows for file listing", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var f m.FileListing
		// Note that we do nothing for URL here - that comes later.
		err := rows.Scan(&f.Size, &f.DateUploaded, &f.ShortURI, &f.LongURI, &f.MimeType,
			&f.ExtraInfo, &f.Deleted)
		if err != nil {
			log.Println("Error retrieving data for file listing", err)
		}
		f.URL = b.DomainName + "/" + f.ShortURI
		listings = append(listings, f)
	}
	err = rows.Err()
	if err != nil {
		log.Println("Error with file listing db rows", err)
	}

	response := ListingResponse{listings, true}
	json.NewEncoder(w).Encode(response)
}

func (b Backend) GetFile(w http.ResponseWriter, r *http.Request) {
	file_uri := mux.Vars(r)["fileid"]
	// TODO show a details page on longuri

	var path string
	var mimetype string
	err := b.DB.QueryRow("select mimetype, path from file where shorturi = ? and deleted = 0",
		file_uri).Scan(&mimetype, &path)
	// TODO say if it's been deleted
	if err != nil {
		if err == sql.ErrNoRows {
			// TODO nice 404 page
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 not found"))
			return
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("error: %s", err)))
			return
		}
	}
	path = b.FileRoot + path
	log.Println("path being getted:", path)
	mimetype = safemime.SafeMime()(mimetype)
	log.Println("safe mime type", mimetype)
	// TODO log unknown mime types here
	w.Header().Set("Content-Type", mimetype)
	http.ServeFile(w, r, path)
}

func (b Backend) RegisterUser(w http.ResponseWriter, r *http.Request) {
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
	if pass != password {
		fail(w, "Wrong admin password")
		return
	}

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

	err = os.Mkdir(fmt.Sprintf("%s/%d", b.FileRoot, uid), os.ModeDir|0777)
	if err != nil {
		fail(w, fmt.Sprintf("Error making user dir for user ID %d: %s", uid, err))
		log.Println("Error making user dir for ID %d: %s", uid, err)
	}

	// Success!
	things := map[string]string{"success": "true",
		"userid": fmt.Sprintf("%d", uid), "token": token}
	json.NewEncoder(w).Encode(things)
}

func (b Backend) UploadFile(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s\n%s", r, r.Body)

	// We need to extract our keys from the payload
	// Token, upload content, mimetype, extra info
	// We could save the original filename, but throw this away for privacy
	// No file size checks here - talk to people if they use lots.

	file := m.File{}
	file.UploadTime = time.Now()

	token := r.FormValue("token")
	if token == "" {
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
			fail(w, fmt.Sprintf("user id db error: %s", err))
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
		err = b.DB.QueryRow("select id from file where ShortURI like ?",
			file.ShortUri+".%").Scan(&id)
		if err == sql.ErrNoRows {
			// Great!
			non_duplicate = true
		} else if err != nil {
			fail(w, fmt.Sprintf("random generation error: %s", err))
			log.Println("Error getting shortname duplicity", err)
			return
		}
	}
	file.ShortUri += urlgen.GetExtension(file.Mime)

	// Save path without the root before it
	file.Path = fmt.Sprintf("%d/%s", file.Uploader, file.ShortUri)

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
		err = b.DB.QueryRow("select id from file where LongURI like ?",
			file.LongUri+".%").Scan(&id)
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
	log.Println("Full path", fullpath)
	f, err := os.Create(fullpath)
	if err != nil {
		log.Printf("Error making disk file on upload: %s", err)
		fail(w, fmt.Sprintf("Error saving file: %s", err))
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
		fail(w, fmt.Sprintf("Could not save file: %s", err))
		log.Printf("Could not save file to DB: %s", err)
		return
	}
	// Finally, success.
	resp := map[string]string{"success": "true",
		"url": fmt.Sprintf("%s/%s", b.DomainName, file.ShortUri)}
	json.NewEncoder(w).Encode(resp)
}

func (b Backend) DeleteFile(w http.ResponseWriter, r *http.Request) {
	fileid := mux.Vars(r)["fileid"]
	token := r.FormValue("token")

	// check that FileID exists, then check that the owner of this token uploaded it.

	// TODO what about LONGURIs?
	var uploaderID int
	err := b.DB.QueryRow("select uploaderID from file where ShortUri = ?",
		fileid).Scan(&uploaderID)
	if err == sql.ErrNoRows {
		fail(w, "No file with this ShortURI")
		return
	} else if err != nil {
		fail(w, "Error checking db for fileid deletion")
		log.Println("Error checking db for fileid for deletion", fileid)
		return
	}

	var token_holder int
	err = b.DB.QueryRow("select ID from user where UploadToken = ?",
		token).Scan(&token_holder)
	if err == sql.ErrNoRows {
		// This token is invalid, no user associated with it.
		fail(w, "invalid token")
		return
	} else if err != nil {
		fail(w, fmt.Sprintf("error checking token: %s", err))
		log.Println("Error checking token for file deletion", err)
		return
	}

	if token_holder != uploaderID {
		fail(w, "you do not have permission to delete this file")
		return
	}

	var deleted int
	err = b.DB.QueryRow("select Deleted from File where ShortURI = ?",
		fileid).Scan(&deleted)
	if err != nil {
		log.Println("Error checking if file deleted with shorturi", fileid)
		fail(w, "Error checking if file already deleted")
		return
	}
	if deleted == 1 {
		fail(w, "Cannot delete file which has already been deleted")
		return
	}

	_, err = b.DB.Exec("UPDATE File SET Deleted = 1 WHERE ShortURI = ?", fileid)
	if err != nil {
		fail(w, "could not delete file entry")
		log.Println("Error setting file as deleted with ShortURI", fileid)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"success": "true"})
}

func (b Backend) createTables() {
	//setup the DB
	_, err := b.DB.Exec(`CREATE TABLE IF NOT EXISTS User (
						   ID INTEGER PRIMARY KEY AUTOINCREMENT,
						   Name TEXT NOT NULl,
						   UploadToken TEXT NOT NULL)`)
	// Token could be UNIQUE but we enforce that on generation anyway
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
						   UploadTime INTEGER NOT NULL,
						   Deleted INTEGER NOT NULL DEFAULT 0)`)
	if err != nil {
		log.Fatal("Could not create table file: ", err)
	}
}
