package db

import (
	"log"

	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func OpenSQLite(path string) *gorm.DB {
	d, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	return d
}
