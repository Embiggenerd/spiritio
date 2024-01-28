package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseName string
	Addr         string
	LogFileName  string
}

func GetConfig() *Config {
	var goEnv = flag.String("go_env", "dev", "which environment")
	flag.Parse()
	fmt.Println("****", *goEnv)
	err := godotenv.Load("pkg/config/" + *goEnv + ".env")
	if err != nil {
		// log.Println(err)
		// log.Fatal("Error loading .env file")

	}

	// // Time allowed to write a message to the peer.
	// writeWait = 10 * time.Second

	// // Time allowed to read the next pong message from the peer.
	// pongWait = 60 * time.Second

	// // Send pings to peer with this period. Must be less than pongWait.
	// pingPeriod = (pongWait * 9) / 10

	// // Maximum message size allowed from peer.
	// maxMessageSize = 512

	// websocketCfg := config{
	// 	writeWait: 10 * time.Second	}
	// hi := os.Getenv("hi")
	// log.Println("&&&", hi)
	// secretKey := os.Getenv("SECRET_KEY")
	cfg := Config{
		DatabaseName: os.Getenv("databasename"),
		Addr:         os.Getenv("addr"),
		LogFileName:  os.Getenv("logfilename"),
	}
	fmt.Println("$^$^$^", cfg.Addr, cfg.DatabaseName, cfg.LogFileName)
	return &cfg
}

type websocketConfig struct {
	// Time allowed to write a message to the peer.
	writeWait time.Duration

	// Time allowed to read the next pong message from the peer.
	pongWait time.Duration

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod time.Duration

	// Maximum message size allowed from peer.
	maxMessageSize int
}
