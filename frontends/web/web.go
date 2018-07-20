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
	payload.Add("token", token) // TODO secret handling
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
	// var uploads []m.FileListing
	if data["success"] == true {
		// for _, el := range data["uploads"] {
			// upload := m.FileListing{el["url"], el["size"], el["date-uploaded"],
				// el["shorturi"], el["longuri"], el["mimetype"], el["extra-info"], el["deleted"]}
		// }
		// log.Println(data["uploads"])
		// uploads, ok := data["uploads"].([]m.FileListing)
		// if !ok {
			// fail(w, "Could not parse response")
			// log.Println("Error converting response uploads to FileListings: ")
			// return
		// }
		// log.Println(uploads)
		tmpl.Execute(w, data["uploads"])
	} else {
		fail(w, data["reason"].(string))
	}

	// err = tmpl.Execute(os.Stdout, data)
	// if err != nil {
	// fail("Could not generate file listing")
	// fmt.Println("Error getting json from backend", err)
	// return
	// }

}
