package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("db.sqlite3"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	err = db.AutoMigrate(&FileLink{})
	if err != nil {
		panic("failed to migrate database")
	}
	return db
}

func AddFileLink(db *gorm.DB, fileLink *FileLink) error {
	res := db.Create(&fileLink)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func GetFileLink(db *gorm.DB, link string) (*FileLink, error) {
	var fileLink FileLink
	res := db.Model(FileLink{}).Where("link = ?", link).First(&fileLink)
	if res.Error != nil {
		return nil, res.Error
	}
	return &fileLink, nil
}

func ExpireFileLink(db *gorm.DB, link string) error {
	res := db.Model(FileLink{}).Where("link = ?", link).Update("expired", true)
	if res.Error != nil {
		return res.Error
	}
	return nil
}
