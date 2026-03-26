package main

import (
	"log"
	"os"

	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/api"
	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/config"
)

func main() {
	configPath := os.Getenv("NANOCPA_CONFIG")
	if configPath == "" {
		configPath = "config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	server := api.NewServer(cfg)
	if err := server.Run(); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
