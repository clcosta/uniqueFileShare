package main

import (
	"github.com/clcosta/uniqueShareFile/src/background"
	"github.com/clcosta/uniqueShareFile/src/database"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db := database.GetDB()
	worker := background.NewWorker()
	worker.Run()

	router := mux.NewRouter()
	router.HandleFunc("/", homePage).Methods("GET")
	router.HandleFunc("/success/{link}", successUploadPage(db)).Methods("GET")
	router.HandleFunc("/upload", uploadForm(db, worker)).Methods("POST")
	router.HandleFunc("/d/{link}", downloadFile(db, worker)).Methods("GET")

	log.Println("Running server at port:", port)
	http.ListenAndServe(":"+port, router)
}
