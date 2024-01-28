package db

import (
	"context"
	"fmt"

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
	fmt.Println("cfg.DatabaseName" + cfg.DatabaseName)
	db, err := gorm.Open(sqlite.Open("pkg/db/data/"+cfg.DatabaseName), &gorm.Config{})
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Info("Database is connected")
	database := &Database{
		DB: db,
	}
	return database, err
}
