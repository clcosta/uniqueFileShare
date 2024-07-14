package server

import (
	"fmt"
	"github.com/clcosta/uniqueShareFile/pkg/background"
	"github.com/clcosta/uniqueShareFile/pkg/fileHandler"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/clcosta/uniqueShareFile/internal/database"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

var FileHandler = fileHandler.NewFileHandler()

const TimeToDeleteFileInSeconds = 60 * 5 // 5 minutes
const FileSizeLimit = 2 << 20            // 2MB

func homePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./public/index.html")
}

func successUploadPage(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		link := vars["link"]
		var fileLink database.FileLink

		res := db.Where("link = ?", link).First(&fileLink)
		if res.Error != nil {
			http.Error(w, res.Error.Error(), http.StatusInternalServerError)
			return
		}
		if res.RowsAffected == 0 {
			http.Error(w, "Link not found", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, "./public/success.html")
	}
}

func uploadForm(db *gorm.DB, bgWorker *background.Worker) http.HandlerFunc {
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

		res := db.Create(&fileLink)
		if res.Error != nil {
			http.Error(w, res.Error.Error(), http.StatusInternalServerError)
			return
		}
		job := background.Job{Identifier: link, Job: FileHandler.DeleteFile, Args: []interface{}{filePath, TimeToDeleteFileInSeconds}}
		bgWorker.AddJob(job)

		successUrl := fmt.Sprintf("/success/%s", link)
		http.Redirect(w, r, successUrl, http.StatusSeeOther)
	}
}

func downloadFile(db *gorm.DB, bgWorker *background.Worker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		link := vars["link"]
		var fileLink database.FileLink

		res := db.Where("link = ?", link).First(&fileLink)
		if res.Error != nil {
			http.Error(w, res.Error.Error(), http.StatusInternalServerError)
			return
		}
		if res.RowsAffected == 0 {
			http.Error(w, "Link not found", http.StatusNotFound)
			return
		}

		if fileLink.Expired || fileLink.ExpiresAt.Before(time.Now()) {
			fileLink.Expired = true
			db.Save(&fileLink)

			err := FileHandler.DeleteFile(fileLink.PathToFile)
			bgWorker.RemoveJob(link)
			if err != nil {
				log.Println(err.Error())
			}

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
