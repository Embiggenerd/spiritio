package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/Embiggenerd/spiritio/pkg/chat"
	"github.com/Embiggenerd/spiritio/pkg/config"
)

var addr = flag.String("addr", ":8080", "http service address")

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	fmt.Println(path)
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)

	fmt.Println(basepath)

	http.ServeFile(w, r, "../../static/index.html")
}

func main() {
	flag.Parse()
	cfg := config.GetConfig()
	// websocketService := chat.NewWebsocketService(cfg)
	websocketService := chat.NewWebsocketService(cfg)
	go websocketService.Run()
	// websocketClient := chat.NewWebsocketClient(websocketService, cfg)
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		chat.ServeWs(websocketService, w, r)
	})
	server := &http.Server{
		Addr:              *addr,
		ReadHeaderTimeout: 3 * time.Second,
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
