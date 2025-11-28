package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         	string 			`yaml:"env" env-default:"local"`
	StoragePath 	string 			`yaml:"storage_path" env-required:"true"`
	HTTPServer	 					`yaml:"http_server"`
}

type HTTPServer struct {
	Addres  		string			`yaml:"address" env-default:"0.0.0.0:8080"`
	Timeout			time.Duration	`yaml:"timeout" env-default:"5s"`
	IdleTimeout		time.Duration	`yaml:"idle_timeout" env-default:"60s"`

}

type Images struct {
	AllowedFormats	[]string		`yaml:"allowed_formats"`
	MaxUploadSizeMb int64			`yaml:"max_upload_size_mb" env-default:"10"`	
}

func MustLoad() *Config{
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("error opening config file: %s", err)
	}

	var cfg Config

	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatalf("error read config file: %s", err)
	}

	return &cfg
}