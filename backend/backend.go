package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/expectocode/oryza/backend/models"
	"github.com/gorilla/mux"
)

func main() {
	db_path := os.Getenv("ORYZA_DB")
	log.Println("Oryza db path: ", db_path)
	b := models.Setup(db_path)

	router := mux.NewRouter()
	router.HandleFunc("/api/upload", b.UploadFile).Methods("POST")
	router.HandleFunc("/api/register", b.RegisterUser).Methods("POST")
	router.HandleFunc("/api/list-uploads", b.ListUploads).Methods("GET")
	router.HandleFunc("/{fileid}", b.DeleteFile).Methods("DELETE")
	router.HandleFunc("/{fileid}", b.GetFile).Methods("GET")

	// TODO add file listing
	// TODO add token existence verification
	// TODO detail-related api calls

	srv := &http.Server{
		Handler:      router,
		Addr:         ":443",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServeTLS("tls/server.crt", "tls/server.key"))
}
