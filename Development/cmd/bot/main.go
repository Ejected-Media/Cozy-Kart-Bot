package main

import (
	"embed"
	"log"
	"os"
	"os/signal"
	"syscall"

	"cozy-kart/internal/bot"
	"cozy-kart/internal/store"
	"cozy-kart/internal/web"
)

// Embed the 'web' folder from the project root into the binary
//go:embed ../../web/*
var content embed.FS

func main() {
	// 1. Config
	token := os.Getenv("DISCORD_TOKEN")
	twitchUser := os.Getenv("TWITCH_CHANNEL")
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./cozy-kart.db"
	}

	// 2. Initialize Database
	db, err := store.New(dbPath)
	if err != nil {
		log.Fatalf("Database error: %v", err)
	}

	// 3. Start Web Server (in background)
	go web.StartServer(db, content, twitchUser)

	// 4. Start Discord Bot (in background)
	if token != "" {
		dg := bot.Start(token, db)
		defer dg.Close()
	} else {
		log.Println("⚠️ DISCORD_TOKEN missing. Running in Web-Only mode.")
	}

	// 5. Keep Alive
	log.Println("System running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
