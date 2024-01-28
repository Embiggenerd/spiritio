package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/Embiggenerd/spiritio/pkg/config"
	"github.com/Embiggenerd/spiritio/pkg/db"
	"github.com/Embiggenerd/spiritio/pkg/logger"
	"github.com/Embiggenerd/spiritio/pkg/rooms"
	"github.com/Embiggenerd/spiritio/pkg/server/handlers"
	"github.com/Embiggenerd/spiritio/pkg/utils"
)

var addr = flag.String("addr", ":8080", "http service address")

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "../static/index.html")
}

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(utils.WithMetadata(context.Background()))
	defer cancel()

	cfg := config.GetConfig()
	logger := logger.NewLoggerService(ctx, cfg)
	db, err := db.Init(ctx, cfg, logger)
	rooms := rooms.NewRoomsService(ctx, cfg, db)

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handlers.ServeWs(rooms, w, r)
	})

	server := &http.Server{
		Addr:              *addr,
		ReadHeaderTimeout: 3 * time.Second,
	}
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
