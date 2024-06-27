package main

import (
	"fmt"
	"github.com/clcosta/uniqueShareFile/src/database"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func saveFile(filepath string, file multipart.File) error {
	dst, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return err
	}
	return nil
}

func deleteFile(filepath string) error {
	return os.Remove(filepath)
}

func homePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "src/assets/index.html")
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
		http.ServeFile(w, r, "src/assets/success.html")
	}
}

func uploadForm(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(2 << 20) // Define limit to 2MB
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
		saveFile(filePath, file)
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
		successUrl := fmt.Sprintf("/success/%s", link)
		http.Redirect(w, r, successUrl, http.StatusSeeOther)
	}
}

func downloadFile(db *gorm.DB) http.HandlerFunc {
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

			deleteFile(fileLink.PathToFile)

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

		err = deleteFile(fileLink.PathToFile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}
}
