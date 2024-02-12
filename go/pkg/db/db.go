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

func Init(ctx context.Context, cfg *config.Config, log logger.Logger) *Database {
	db, err := gorm.Open(sqlite.Open("pkg/db/data/"+cfg.DatabaseName), &gorm.Config{})
	if err != nil {
		log.Fatal(err.Error())
	}
	database := &Database{
		DB: db,
	}
	log.Info("database is initialized")
	return database
}
