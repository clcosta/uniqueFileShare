package database

import (
	"gorm.io/gorm"
	"time"
)

type FileLink struct {
	gorm.Model
	ID         uint
	Link       string
	PathToFile string
	Expired    bool
	ExpiresAt  time.Time
}
