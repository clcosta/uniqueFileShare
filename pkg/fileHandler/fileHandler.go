package fileHandler

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

type FileHandler struct{}

func NewFileHandler() *FileHandler {
	return &FileHandler{}
}

func (fh *FileHandler) SaveFile(filePathStr string, file multipart.File) (string, error) {
	dir := filepath.Dir(filePathStr)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	if _, err := os.Stat(filePathStr); err == nil {
		ext := filepath.Ext(filePathStr)
		base := filepath.Base(filePathStr)
		name := base[0 : len(base)-len(ext)]
		timestamp := time.Now().Format("20060102150405")
		filePathStr = fmt.Sprintf("%s/%s_%s%s", dir, name, timestamp, ext)
	}

	dst, err := os.Create(filePathStr)
	if err != nil {
		return "", err
	}
	defer func(dst *os.File) {
		err := dst.Close()
		if err != nil {
			log.Println("Error closing file", err)
		}
	}(dst)

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}
	return filePathStr, nil
}

func (fh *FileHandler) DeleteFile(args ...interface{}) error {
	filePathStr := args[0].(string)
	var timeToWait int
	if len(args) > 1 {
		timeToWait = args[1].(int)
	} else {
		timeToWait = 0
	}
	log.Println("Deleting file", filePathStr, "in", timeToWait, "seconds")
	time.Sleep(time.Duration(timeToWait) * time.Second)
	err := os.Remove(filePathStr)
	log.Println("File", filePathStr, "deleted")
	return err
}
