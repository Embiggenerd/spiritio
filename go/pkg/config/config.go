package config

import (
	"flag"
	"log"
	"os"
	"reflect"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseName       string `default:"dev.db"`
	Addr               string `default:":8080"`
	LogFileName        string `default:"dev.log"`
	AccessTokenSecret  string `default:"our_secret"`
	MaxPeerConnections int    `default:"4"`
}

func GetConfig() *Config {
	var goEnv = flag.String("go_env", "dev", "which environment")
	flag.Parse()

	err := godotenv.Load("pkg/config/" + *goEnv + ".env")
	if err != nil {
		log.Println(err)
	}

	cfg := Config{
		DatabaseName: os.Getenv("databasename"),
		Addr:         os.Getenv("addr"),
		LogFileName:  os.Getenv("logfilename"),
	}
	const tagName = "default"

	t := reflect.TypeOf(cfg)

	for i := 0; i < t.NumField(); i++ {
		// Get the field
		field := t.Field(i)

		// Get the field tag value
		tag := field.Tag.Get(tagName)

		// Get the value
		r := reflect.ValueOf(cfg)
		v := reflect.Indirect(r).FieldByName(field.Name)

		// If value is invalid, use default value from tag
		if v.String() == "" {
			o := reflect.Indirect(reflect.ValueOf(&cfg))
			o.FieldByName(field.Name).SetString(tag)
		}
	}

	return &cfg
}
