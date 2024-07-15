package server

import (
	"log"
	"net/http"
	"os"

	"github.com/clcosta/uniqueShareFile/internal/database"
	"github.com/clcosta/uniqueShareFile/pkg/background"
	"github.com/gorilla/mux"
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
	router.HandleFunc("/", homePageHandler).Methods("GET")
	router.HandleFunc("/success/{link}", successUploadPageHandler(db)).Methods("GET")
	router.HandleFunc("/upload", uploadFormHandler(db, worker)).Methods("POST")
	router.HandleFunc("/d/{link}", downloadFileHandler(db, worker)).Methods("GET")

	log.Println("Running server at port:", port)
	http.ListenAndServe(":"+port, router)
}
