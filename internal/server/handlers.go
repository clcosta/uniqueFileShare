package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/clcosta/uniqueShareFile/pkg/background"
	"github.com/clcosta/uniqueShareFile/pkg/fileHandler"

	"github.com/clcosta/uniqueShareFile/internal/database"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

var FileHandler = fileHandler.NewFileHandler()

const TimeToDeleteFileInSeconds = 60 * 5 // 5 minutes
const FileSizeLimit = 20 << 20           // 20MB

type Stats struct {
	Link             string
	Downloaded       bool
	ExpiresInSeconds int
}

func homePageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./public/index.html")
}

func successUploadPageHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		link := vars["link"]
		fileLink, err := database.GetFileLink(db, link)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if fileLink == nil {
			http.Error(w, "Link not found", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, "./public/success.html")
	}
}

func uploadFormHandler(db *gorm.DB, bgWorker *background.Worker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(FileSizeLimit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if handler.Size > FileSizeLimit {
			http.Error(w, "File size is greater than 20MB", http.StatusUnprocessableEntity)
			return
		}

		defer file.Close()

		filePath := filepath.Join("tmp", handler.Filename)
		filePath, err = FileHandler.SaveFile(filePath, file)
		if err != nil {
			log.Println("Error saving file", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		link := uuid.New().String()

		fileLink := database.FileLink{
			Link:       link,
			PathToFile: filePath,
			Expired:    false,
			ExpiresAt:  time.Now().Add(TimeToDeleteFileInSeconds * time.Second),
		}

		err = database.AddFileLink(db, &fileLink)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		job := background.Job{
			Identifier: link,
			Job:        FileHandler.DeleteFile,
			Args:       []interface{}{filePath, TimeToDeleteFileInSeconds},
			CallBack: func() error {
				err := database.ExpireFileLink(db, link)
				if err != nil {
					return err
				}
				bgWorker.RemoveJob(link)
				return nil
			},
		}
		bgWorker.AddJob(job)

		successUrl := fmt.Sprintf("/success/%s", link)
		http.Redirect(w, r, successUrl, http.StatusSeeOther)
	}
}

func downloadPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./public/download.html")
}

func expiredPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./public/expired.html")
}

func downloadFileHandler(db *gorm.DB, bgWorker *background.Worker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		link := vars["link"]

		fileLink, err := database.GetFileLink(db, link)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if fileLink == nil {
			http.Error(w, "Link not found", http.StatusNotFound)
			return
		}

		if fileLink.Expired || fileLink.ExpiresAt.Before(time.Now()) {
			if !fileLink.Expired {
				fileLink.Expired = true
				db.Save(&fileLink)
				err := FileHandler.DeleteFile(fileLink.PathToFile)
				if err != nil {
					log.Println(err.Error())
				}
			}
			bgWorker.RemoveJob(link)
			http.Redirect(w, r, "/expired", http.StatusPermanentRedirect)
			return
		}
		file, err := os.Open(fileLink.PathToFile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(fileLink.PathToFile))
		w.Header().Set("Content-Type", "application/octet-stream")
		http.ServeContent(w, r, filepath.Base(fileLink.PathToFile), time.Now(), file)
		file.Close()

		fileLink.Expired = true
		db.Save(&fileLink)

		err = FileHandler.DeleteFile(fileLink.PathToFile)
		bgWorker.RemoveJob(link)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}
}

func statsHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		link := vars["link"]

		fileLink, err := database.GetFileLink(db, link)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if fileLink == nil {
			http.Error(w, "Link not found", http.StatusNotFound)
			return
		}
		expiresInSeconds := int(time.Until(fileLink.ExpiresAt).Seconds())
		if expiresInSeconds < 0 {
			expiresInSeconds = 0
		}
		stats := Stats{
			Link:             fileLink.Link,
			Downloaded:       fileLink.Expired,
			ExpiresInSeconds: expiresInSeconds,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(stats)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
