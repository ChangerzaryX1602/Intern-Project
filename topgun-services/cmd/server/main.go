package main

import (
	"flag"
	"log"
	"os"

	_ "topgun-services/docs"
	server "topgun-services/internal/infrastructure"
	"topgun-services/pkg/config"
)

var (
	version string
	build   string
	runEnv  string
)

func init() {
	// read running flag
	if len(os.Getenv("ENV")) != 0 {
		runEnv = os.Getenv("ENV")
	} else {
		flagEnv := flag.String("env", "dev", "A config file name without .env")
		flag.Parse()
		runEnv = *flagEnv
	}

	// load config by running flag
	if err := config.LoadConfig(runEnv); err != nil {
		log.Fatalf("error while loading the env:\n %+v", err)
	}
}

// @title KKU GS ADMISSION SERVICES API
// @schemes http https
// @securityDefinitions.apiKey ApiKeyAuth
// @in header
// @name Authorization
func main() {

	// init server
	server, err := server.NewServer(version, build, runEnv)
	if err != nil {
		log.Fatalf("error while create server:\n %+v", err)
	}

	server.Run()
}
