package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"
)

var addr = flag.String("addr", ":8081", "http service address")

func serveHome(w http.ResponseWriter, r *http.Request) {
	Data := `{"hello":"world"}`
	jData, err := json.Marshal(Data)
	if err != nil {
		// handle error
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jData)
}

func main() {
	flag.Parse()
	http.HandleFunc("/", serveHome)

	server := &http.Server{
		Addr:              *addr,
		ReadHeaderTimeout: 3 * time.Second,
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	log.Printf("User service up on port %s", *addr)
}
