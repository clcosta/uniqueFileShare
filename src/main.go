package main

import (
	"fmt"
	"github.com/clcosta/uniqueShareFile/src/database"
	"github.com/gorilla/mux"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db := database.GetDB()

	router := mux.NewRouter()
	router.HandleFunc("/", homePage).Methods("GET")
	router.HandleFunc("/success/{link}", successUploadPage(db)).Methods("GET")
	router.HandleFunc("/upload", uploadForm(db)).Methods("POST")
	router.HandleFunc("/d/{link}", downloadFile(db)).Methods("GET")

	fmt.Println("Running server at port:", port)
	http.ListenAndServe(":"+port, router)
}
