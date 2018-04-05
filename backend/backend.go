package main

import (
	"net/http"
	"os"
	"time"
	"log"

	"github.com/gorilla/mux"
	"github.com/expectocode/oryza/backend/models"
)

func main() {
	db_path := os.Getenv("ORYZA_DB")
	log.Println("Oryza db path: ", db_path)
	b := models.NewBackend(db_path)

	router := mux.NewRouter()
	router.HandleFunc("/api/upload", b.UploadFile).Methods("POST")
	router.HandleFunc("/api/register", b.RegisterUser).Methods("POST")
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
