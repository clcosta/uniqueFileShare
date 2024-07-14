package server

import (
	"github.com/clcosta/uniqueShareFile/internal/database"
	"github.com/clcosta/uniqueShareFile/pkg/background"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

func RunServer() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db := database.GetDB()

	worker := background.NewWorker()
	worker.Run()

	router := mux.NewRouter()
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./public/"))))
	router.HandleFunc("/", homePage).Methods("GET")
	router.HandleFunc("/success/{link}", successUploadPage(db)).Methods("GET")
	router.HandleFunc("/upload", uploadForm(db, worker)).Methods("POST")
	router.HandleFunc("/d/{link}", downloadFile(db, worker)).Methods("GET")

	log.Println("Running server at port:", port)
	http.ListenAndServe(":"+port, router)
}
