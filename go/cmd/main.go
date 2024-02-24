package main

import (
	"context"
	"net/http"

	"github.com/Embiggenerd/spiritio/pkg/config"
	"github.com/Embiggenerd/spiritio/pkg/db"
	"github.com/Embiggenerd/spiritio/pkg/logger"
	"github.com/Embiggenerd/spiritio/pkg/rooms"
	"github.com/Embiggenerd/spiritio/pkg/server"
	"github.com/Embiggenerd/spiritio/pkg/users"
	"github.com/Embiggenerd/spiritio/pkg/utils"
)

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "static/index.html")
}

func main() {
	ctx, cancel := context.WithCancel(utils.WithMetadata(context.Background()))
	defer cancel()

	cfg := config.GetConfig()
	logger := logger.NewLoggerService(ctx, cfg)
	db := db.Init(ctx, cfg, logger)
	roomsService := rooms.NewRoomsService(ctx, cfg, logger, db)
	usersService := users.NewUsersService(ctx, cfg, logger, db)
	apiServer := server.NewServer(ctx, cfg, logger, roomsService, usersService)

	apiServer.Run()
}
