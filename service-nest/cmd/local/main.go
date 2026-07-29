package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	appconfig "service-nest/config"
	"service-nest/internal/app"
	"service-nest/util"
)

func main() {
	loadDotEnv()

	cfg, err := appconfig.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	if err := util.InitJWT(cfg.JWTSecret); err != nil {
		log.Fatalf("jwt init failed: %v", err)
	}

	application, err := app.Bootstrap(cfg)
	if err != nil {
		log.Fatalf("bootstrap failed: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	log.Printf("service-nest local server starting on %s (%s)", addr, cfg.String())
	if err := http.ListenAndServe(addr, application.Router); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func loadDotEnv() {
	candidates := []string{
		".env",
		filepath.Join("..", ".env"),
		filepath.Join("..", "..", ".env"),
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		if err := godotenv.Load(path); err != nil {
			log.Printf("warning: failed to load %s: %v", path, err)
			continue
		}
		log.Printf("loaded environment from %s", path)
		return
	}

	log.Printf("no .env file found; using process environment only")
}
