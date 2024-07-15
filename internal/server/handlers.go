package server

import (
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
const FileSizeLimit = 2 << 20            // 2MB

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
		err := r.ParseMultipartForm(FileSizeLimit) // Define limit to 2MB
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
			http.Error(w, "File size is greater than 2MB", http.StatusUnprocessableEntity)
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
			ExpiresAt:  time.Now().Add(1 * time.Hour),
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
				return database.ExpireFileLink(db, link)
			},
		}
		bgWorker.AddJob(job)

		successUrl := fmt.Sprintf("/success/%s", link)
		http.Redirect(w, r, successUrl, http.StatusSeeOther)
	}
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
			http.Error(w, "Link expired", http.StatusUnprocessableEntity)
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
