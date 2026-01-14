Here is the complete source code broken down into the professional file structure we designed. This separates your concerns (Database vs. Bot vs. Website) making it easy to maintain.
📂 The Directory Structure
First, create these folders on your computer:
cozy-kart/
├── cmd/
│   └── bot/
│       └── main.go        # Entry point
├── internal/
│   ├── store/
│   │   └── store.go       # Database logic
│   ├── bot/
│   │   └── bot.go         # Discord logic
│   └── web/
│       └── server.go      # Website logic
├── web/
│   └── templates/
│       └── index.html     # The Dashboard UI
├── Dockerfile
├── docker-compose.yml
└── go.mod

1. go.mod
Run go mod init cozy-kart and go mod tidy to generate this, or copy it:
module cozy-kart

go 1.23

require (
	github.com/bwmarrin/discordgo v0.27.1
	modernc.org/sqlite v1.28.0
)

2. internal/store/store.go
This handles all SQLite interactions.
package store

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	DB *sql.DB
}

type Player struct {
	DiscordID   string
	MKTUsername string
	AvatarURL   string
	Status      string
	JoinedAt    time.Time
}

func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Create tables if they don't exist
	schema := `
	CREATE TABLE IF NOT EXISTS players (
		discord_id TEXT PRIMARY KEY,
		mkt_username TEXT,
		avatar_url TEXT
	);
	CREATE TABLE IF NOT EXISTS queue (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		discord_id TEXT,
		status TEXT DEFAULT 'QUEUED', -- QUEUED, INVITED
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}

	return &Store{DB: db}, nil
}

func (s *Store) AddToQueue(discordID, mktName, avatar string) error {
	// 1. Upsert Player (Update name/avatar if they changed)
	_, err := s.DB.Exec(`
		INSERT INTO players (discord_id, mkt_username, avatar_url) VALUES (?, ?, ?)
		ON CONFLICT(discord_id) DO UPDATE SET mkt_username=excluded.mkt_username, avatar_url=excluded.avatar_url`,
		discordID, mktName, avatar)
	if err != nil {
		return err
	}

	// 2. Check if already in queue
	var count int
	s.DB.QueryRow("SELECT COUNT(*) FROM queue WHERE discord_id = ? AND status = 'QUEUED'", discordID).Scan(&count)
	if count > 0 {
		return fmt.Errorf("already_queued")
	}

	// 3. Add to Queue
	_, err = s.DB.Exec("INSERT INTO queue (discord_id, status) VALUES (?, 'QUEUED')", discordID)
	return err
}

// GetActiveGrid fetches up to 8 players for the N64 grid
func (s *Store) GetActiveGrid() ([]Player, error) {
	rows, err := s.DB.Query(`
		SELECT p.discord_id, p.mkt_username, p.avatar_url, q.status, q.created_at
		FROM queue q
		JOIN players p ON q.discord_id = p.discord_id
		WHERE q.status IN ('QUEUED', 'INVITED')
		ORDER BY q.created_at ASC
		LIMIT 8
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grid []Player
	for rows.Next() {
		var p Player
		if err := rows.Scan(&p.DiscordID, &p.MKTUsername, &p.AvatarURL, &p.Status, &p.JoinedAt); err != nil {
			log.Println("Row scan error:", err)
			continue
		}
		grid = append(grid, p)
	}
	return grid, nil
}

3. internal/bot/bot.go
This handles the Discord connection and slash commands.
package bot

import (
	"cozy-kart/internal/store"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

func Start(token string, db *store.Store) *discordgo.Session {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	// Register the Interaction Handler
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}

		if i.ApplicationCommandData().Name == "join" {
			handleJoin(s, i, db)
		}
	})

	if err := dg.Open(); err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}

	// Register Slash Command
	cmd := &discordgo.ApplicationCommand{
		Name:        "join",
		Description: "Join the Cozy-Kart Queue",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "username",
				Description: "Your Mario Kart Tour Username",
				Required:    true,
			},
		},
	}
	
	// Register globally (takes 1 hour) or pass Guild ID for instant update
	_, err = dg.ApplicationCommandCreate(dg.State.User.ID, "", cmd)
	if err != nil {
		log.Printf("Cannot create command: %v", err)
	}

	log.Println("🤖 Bot is connected!")
	return dg
}

func handleJoin(s *discordgo.Session, i *discordgo.InteractionCreate, db *store.Store) {
	mktName := i.ApplicationCommandData().Options[0].StringValue()
	user := i.Member.User
	avatar := user.AvatarURL("128") // Get small avatar

	err := db.AddToQueue(user.ID, mktName, avatar)
	
	msg := fmt.Sprintf("🏎️ **%s** joined the grid as *%s*!", user.Username, mktName)
	if err != nil && err.Error() == "already_queued" {
		msg = "✋ You are already on the grid!"
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: msg},
	})
}

4. internal/web/server.go
This serves the HTML and injects the data.
package web

import (
	"cozy-kart/internal/store"
	"embed"
	"html/template"
	"log"
	"net/http"
)

// We pass the filesystem in from main.go to keep this package clean
func StartServer(db *store.Store, fs embed.FS, twitchUser string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		grid, _ := db.GetActiveGrid()

		// Fill empty slots if less than 8 players (for visual grid)
		displayGrid := make([]store.Player, 8)
		for i := 0; i < 8; i++ {
			if i < len(grid) {
				displayGrid[i] = grid[i]
			} else {
				// Empty placeholder
				displayGrid[i] = store.Player{MKTUsername: "WAITING...", Status: "EMPTY"}
			}
		}

		data := struct {
			TwitchUser string
			Grid       []store.Player
		}{
			TwitchUser: twitchUser,
			Grid:       displayGrid,
		}

		// Parse template
		tmpl, err := template.ParseFS(fs, "web/templates/index.html")
		if err != nil {
			http.Error(w, "Template error: "+err.Error(), 500)
			return
		}
		tmpl.Execute(w, data)
	})

	log.Println("🌐 Web Dashboard running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

5. web/templates/index.html
The Spectator UI with the Twitch Embed and the N64 Grid.
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Cozy-Kart Live</title>
    <meta http-equiv="refresh" content="10">
    <style>
        body { background-color: #1a1a1a; color: white; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; margin: 0; padding: 20px; }
        .container { max-width: 1200px; margin: 0 auto; text-align: center; }
        
        /* The Twitch Stage */
        #twitch-embed { 
            width: 100%; 
            aspect-ratio: 16/9; 
            border: 4px solid #5865F2; 
            border-radius: 8px; 
            margin-bottom: 20px;
            background: black;
        }

        /* The N64 Grid */
        .grid-container {
            display: grid;
            grid-template-columns: repeat(4, 1fr);
            gap: 15px;
        }

        .player-card {
            background: #2b2b2b;
            border-radius: 10px;
            padding: 10px;
            display: flex;
            align-items: center;
            border: 2px solid #444;
        }
        
        .player-card.active { border-color: #4caf50; background: #1e3a1e; }
        
        .avatar { width: 50px; height: 50px; border-radius: 50%; margin-right: 10px; background: #444; object-fit: cover;}
        .info { text-align: left; }
        .name { font-weight: bold; font-size: 1.1em; }
        .status { font-size: 0.8em; color: #aaa; text-transform: uppercase; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🏁 Cozy-Kart Arena</h1>
        
        <div id="twitch-embed"></div>

        <div class="grid-container">
            {{ range .Grid }}
                {{ if eq .Status "EMPTY" }}
                <div class="player-card">
                    <div class="avatar"></div>
                    <div class="info">
                        <div class="name" style="color:#666">WAITING...</div>
                    </div>
                </div>
                {{ else }}
                <div class="player-card active">
                    <img src="{{ .AvatarURL }}" class="avatar">
                    <div class="info">
                        <div class="name">{{ .MKTUsername }}</div>
                        <div class="status">{{ .Status }}</div>
                    </div>
                </div>
                {{ end }}
            {{ end }}
        </div>
    </div>

    <script src="https://embed.twitch.tv/embed/v1.js"></script>
    <script type="text/javascript">
      new Twitch.Embed("twitch-embed", {
        width: "100%",
        height: "100%",
        channel: "{{ .TwitchUser }}", 
        layout: "video",
        autoplay: true,
        parent: ["localhost", "127.0.0.1"] // Add your domain here when deploying!
      });
    </script>
</body>
</html>

6. cmd/bot/main.go
The glue that wires it all together.
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

7. docker-compose.yml
For running locally.
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data
    environment:
      - DISCORD_TOKEN=your_token_here
      - TWITCH_CHANNEL=your_twitch_username
      - DB_PATH=/app/data/cozy.db

8. Dockerfile
For Google Cloud / Production.
FROM golang:1.23-alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Build the binary inside cmd/bot
RUN go build -o cozy-kart ./cmd/bot

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/cozy-kart .
EXPOSE 8080
CMD ["./cozy-kart"]

