package config

import (
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseName string
	Addr         string
	LogFileName  string
}

func GetConfig() *Config {
	flag.Parse()

	var goEnv = flag.String("go_env", "dev", "which environment")

	err := godotenv.Load("pkg/config/" + *goEnv + ".env")
	if err != nil {
		log.Println(err)
	}

	cfg := Config{
		DatabaseName: os.Getenv("databasename"),
		Addr:         os.Getenv("addr"),
		LogFileName:  os.Getenv("logfilename"),
	}

	return &cfg
}
