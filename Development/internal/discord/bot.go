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
