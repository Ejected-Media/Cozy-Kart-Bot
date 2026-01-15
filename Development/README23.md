Here is the main.go file.
This is the Conductor. In Go architecture, main should be very boring; its only job is to wire things together and keep the lights on.
The critical architectural challenge here is running Two Engines at once:
 * The Discord Bot: A persistent connection to Discord's gateway.
 * The HTTP/WebSocket Server: The API that the Android phones connect to.
They must run in parallel (concurrency) so one doesn't block the other.
cmd/cozy-kart/main.go
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv" // Common library for .env files
	
	// Local imports (assuming your module is named cozy-kart)
	"github.com/yourname/cozy-kart/pkg/api"
	"github.com/yourname/cozy-kart/pkg/bot"
	"github.com/yourname/cozy-kart/pkg/game"
)

func main() {
	// 1. Load Environment Variables (Secrets)
	// We fail fast if the tokens aren't there.
	if err := godotenv.Load(); err != nil {
		log.Println("Info: No .env file found, relying on System ENV variables")
	}

	token := os.Getenv("DISCORD_TOKEN")
	appID := os.Getenv("DISCORD_APP_ID")
	guildID := os.Getenv("DISCORD_GUILD_ID")

	if token == "" {
		log.Fatal("❌ Error: DISCORD_TOKEN is required.")
	}

	// 2. Initialize the Single Source of Truth (The Race Manager)
	// This struct holds the state of the race, the pot, and the racers.
	// It is created ONCE and shared with everyone.
	raceManager := game.NewRaceManager()
	log.Println("🏁 Race Manager initialized.")

	// 3. Start the HTTP Server (For Android Phones)
	// We run this in a goroutine "go func()" so it doesn't block the main thread.
	go func() {
		// Pass the manager to the API so phones can get real data
		apiHandler := api.NewHandler(raceManager)
		
		http.HandleFunc("/ws", apiHandler.HandleWebSocket) // The socket endpoint
		http.HandleFunc("/health", apiHandler.HandleHealth) // Simple ping
		
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		
		log.Printf("🌍 HTTP Server listening on port %s...", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("❌ HTTP Server Failed: %v", err)
		}
	}()

	// 4. Start the Discord Bot (For Admin & Chat)
	// We configure the bot and inject the SAME raceManager instance.
	botConfig := bot.Config{
		Token:   token,
		AppID:   appID,
		GuildID: guildID,
	}

	discordBot, err := bot.New(botConfig, raceManager)
	if err != nil {
		log.Fatalf("❌ Error creating bot: %v", err)
	}

	// 5. Run the Bot (Blocking)
	// This function (from bot.go) contains the "Block Forever" logic.
	// The program stays alive here until we hit CTRL-C.
	log.Println("🤖 Starting CozyBot...")
	if err := discordBot.Start(); err != nil {
		log.Fatalf("❌ Bot crashed: %v", err)
	}
}

Why this structure is powerful:
 * Shared State (raceManager): Both the discordBot and the apiHandler receive the pointer to raceManager.
   * If an Admin types /host bonus 500 in Discord -> The Bot updates raceManager.
   * The raceManager updates its internal state.
   * The apiHandler (WebSocket) sees the change and pushes it to the Android phones instantly.
   * No database reads required for live gameplay. It all happens in memory for speed.
 * Concurrency: The go func() wrapper around the HTTP server is what allows the Web Server and the Discord Bot to coexist in the same application.
 * Simplicity: This file is barely 60 lines long. It delegates all the complex logic to the specialized packages (pkg/bot, pkg/game, pkg/api), keeping the entry point clean.
This concludes the "Backend Core."
We have:
 * bot: Talks to Discord.
 * game: Holds the Logic.
 * main: Starts it all.
The only missing piece mentioned in main.go is pkg/api (the WebSocket handler). Would you like to build that WebSocket Hub next so the phones can actually connect?
