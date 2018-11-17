package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

func fail(w http.ResponseWriter, reason string) {
	w.Write([]byte(fmt.Sprintf("Error generating page: %s", reason)))
	return
}

func UploadsPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("listing.html").ParseFiles("html/listing.html")
	if err != nil {
		panic(err)
	}

	token := mux.Vars(r)["token"]
	if token == "" {
		fail(w, "Token cannot be empty! URL should be of the form https://up.unix.porn/web/{token}/uploads")
		return
	}

	// Make GET request to backend
	payload := url.Values{}
	payload.Add("token", token)
	resp, err := http.Get("https://up.unix.porn/api/list-uploads?" + payload.Encode())
	if err != nil {
		panic(err)
	}

	// Parse response
	defer resp.Body.Close()
	data := make(map[string]interface{})
	rBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fail(w, "Could not generate file listing")
		log.Println("Error getting response from backend", err)
		return
	}
	err = json.Unmarshal(rBody, &data)
	if err != nil {
		fail(w, "Could not generate file listing")
		log.Println("Error getting json from backend", err)
		return
	}

	// is there a better way to put this into the correct struct?
	if data["success"] == true {
		tmpl.Execute(w, data["uploads"])
	} else {
		fail(w, data["reason"].(string))
	}
}

func MainPage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("unix.porn upload service for /rice/ and friends. you can access a list of your files at https://up.unix.porn/web/{your token}/uploads. See https://github.com/expectocode/oryza for details."))
}
