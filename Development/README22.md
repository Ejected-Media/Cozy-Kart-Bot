Here is the source code for the primary bot.go file.
This is the "Engine Room" of the Discord integration. It handles the Lifecycle (Starting up, shutting down) and acts as the bridge between the raw Discord events and our specific logic (Commands and Race Manager).
pkg/bot/bot.go
package bot

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/yourname/cozy-kart/pkg/game" // Import your internal game package
)

// Config holds the secrets and IDs needed to connect.
type Config struct {
	Token   string
	AppID   string // Application ID (for Slash Commands)
	GuildID string // The specific server ID (makes command updates instant during dev)
}

// CozyBot is the main controller for the Discord integration.
// It holds the connection session and a pointer to the Game Logic.
type CozyBot struct {
	Session *discordgo.Session
	Config  Config
	Manager *game.RaceManager // Reference to the backend logic
}

// New creates a fresh instance of the bot but does not connect yet.
func New(cfg Config, raceManager *game.RaceManager) (*CozyBot, error) {
	// Create the Discord Session
	s, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("error creating discord session: %w", err)
	}

	return &CozyBot{
		Session: s,
		Config:  cfg,
		Manager: raceManager,
	}, nil
}

// Start opens the connection, registers handlers, and blocks until a kill signal.
func (b *CozyBot) Start() error {
	
	// 1. Register the "Traffic Cop" (The Interaction Handler)
	// This function (defined in commands.go) decides which logic to run.
	b.Session.AddHandler(b.interactionHandler)

	// 2. Add a handler for when the bot is ready
	b.Session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("🤖 CozyBot is Online! Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
		
		// Optional: Update Status
		s.UpdateGameStatus(0, "Waiting for Race 🏁")
	})

	// 3. Open the Websocket Connection
	// We need intents to see messages if we add non-slash commands later.
	b.Session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages

	err := b.Session.Open()
	if err != nil {
		return fmt.Errorf("error opening connection: %w", err)
	}

	// 4. Register Slash Commands
	// We do this AFTER opening the connection so we can talk to the API.
	b.RegisterCommands()

	log.Println("🟢 Bot is running. Press CTRL-C to exit.")
	
	// 5. Block forever until an OS signal (CTRL-C) is received
	// This prevents the main function from exiting immediately.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	return b.Stop()
}

// Stop gracefully shuts down the connection.
func (b *CozyBot) Stop() error {
	log.Println("🛑 Stopping Bot...")
	
	// Optional: Remove commands on shutdown to clean up? 
	// Usually better to leave them, but strictly for dev you might remove them here.
	
	return b.Session.Close()
}

// AnnounceRaceStart is a public method the RaceManager can call
// to send a message to a specific channel when the race begins.
func (b *CozyBot) AnnounceRaceStart(channelID string) {
	embed := &discordgo.MessageEmbed{
		Title: "🏎️ The Race Has Started!",
		Description: "The queue is closed. Watch the stream to see the results!",
		Color: 0x00E5FF, // Cyan
	}
	b.Session.ChannelMessageSendEmbed(channelID, embed)
}

Key Architectural Notes:
 * b.Session.AddHandler(b.interactionHandler): This is the most important line. It links the low-level Discord events to the high-level logic we wrote in commands.go.
 * Manager *game.RaceManager: The Bot struct "owns" a pointer to the Race Manager. This allows the bot to directly mutate game state (like b.Manager.AddToPot(500)) when a Discord command is received.
 * Graceful Shutdown: The signal.Notify block ensures that if the server restarts, the bot closes the socket cleanly rather than timing out.
Would you like to see the main.go (Application Entry Point) that instantiates this Bot and the Race Manager side-by-side?
