package db

import (
	"context"

	"github.com/Embiggenerd/spiritio/pkg/config"
	"github.com/Embiggenerd/spiritio/pkg/logger"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Database holds our gorm, db instance
type Database struct {
	DB *gorm.DB
}

func Init(ctx context.Context, cfg *config.Config, log logger.Logger) (*Database, error) {
	db, err := gorm.Open(sqlite.Open("dev.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Info("Database is connected")
	database := &Database{
		DB: db,
	}
	return database, err
}
