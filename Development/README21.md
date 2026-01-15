Based on the architecture we have built, the Discord Bot serves two masters: the Racers (Public) and the Race Hosts (Admin).
To make this a complete "Command Interface," we need to expand beyond just the /race command. We need a suite of functions that handle Registration, Status Checks, and Race Management.
Here is the complete Interface Definition for commands.go.
The Command Hierarchy
I have organized the functions into three permission levels:
 * Public: Safe for anyone to use.
 * Racer: Only for registered players (checks database).
 * Admin: Only for the Event Host (requires specific Discord Role).
The Source Code: commands.go
This file handles the registration of commands with Discord and the routing logic (the switch statement) that directs traffic.
package bot

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

// RegisterCommands defines the "API" of our bot.
// It tells Discord: "Here are the buttons and commands I support."
func (b *CozyBot) RegisterCommands() {
	commands := []*discordgo.ApplicationCommand{
		// --- PUBLIC COMMANDS ---
		{
			Name:        "race",
			Description: "🏁 Get your Magic Link to join the dashboard",
		},
		{
			Name:        "stats",
			Description: "📊 View your career earnings and win rate",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "racer",
					Description: "Look up a specific racer (optional)",
					Required:    false,
				},
			},
		},
		{
			Name:        "leaderboard",
			Description: "🏆 Show the Top 5 Racers by Season Earnings",
		},

		// --- ADMIN COMMANDS (Host Only) ---
		{
			Name:        "host",
			Description: "🔧 Admin Tools for the Race Manager",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "open",
					Description: "Open the Green Room for new players",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "start",
					Description: "Lock the queue and start the Ignition Sequence",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "bonus",
					Description: "Inject a Sponsor Bonus into the pot",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "amount",
							Description: "Amount in cents (e.g., 500 = $5.00)",
							Required:    true,
						},
					},
				},
			},
		},
	}

	// Bulk Overwrite (Replaces old commands with these new ones instantly)
	_, err := b.Session.ApplicationCommandBulkOverwrite(b.Config.AppID, b.Config.GuildID, commands)
	if err != nil {
		log.Panicf("Failed to register commands: %v", err)
	}
	log.Println("🤖 Command Interface Successfully Registered.")
}

// Router: This function decides which logic to run based on the user's input.
func (b *CozyBot) interactionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Only handle Slash Commands here
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	switch i.ApplicationCommandData().Name {
	
	case "race":
		b.handleRaceLink(s, i) // The "Magic Link" function we wrote earlier
	
	case "stats":
		b.handleStats(s, i)
	
	case "leaderboard":
		b.handleLeaderboard(s, i)
	
	case "host":
		// Check permissions! Only allow users with "Race Host" role (or Administrator)
		if !hasAdminPermissions(s, i.Member) {
			b.respondError(s, i, "You do not have the keys to the track.")
			return
		}
		b.handleHostSubcommand(s, i)
	}
}

The Handler Logic: handlers.go
Now, let's implement the specific logic for the new commands (stats, leaderboard, and host).
1. Public Command: /stats
This fetches data (simulated here) and displays it as a "Racer Card."
func (b *CozyBot) handleStats(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// 1. Determine target user (Self or Mentioned User)
	targetUser := i.Member.User
	options := i.ApplicationCommandData().Options
	if len(options) > 0 {
		targetUser = options[0].UserValue(s)
	}

	// 2. Mock Database Fetch (Replace with Firestore call later)
	earnings := "$124.50"
	races := 12
	wins := 3

	// 3. Construct the "Racer Card" Embed
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("License: %s", targetUser.Username),
		Color:       0xFFA500, // Cozy Orange
		Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: targetUser.AvatarURL("")},
		Fields: []*discordgo.MessageEmbedField{
			{Name: "💰 Lifetime Earnings", Value: earnings, Inline: true},
			{Name: "🏎️ Races Run", Value: fmt.Sprintf("%d", races), Inline: true},
			{Name: "🏆 Gold Medals", Value: fmt.Sprintf("%d", wins), Inline: true},
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Embeds: []*discordgo.MessageEmbed{embed}},
	})
}

2. Admin Command: /host [action]
This uses Subcommands to group admin tools together. This is where the Bot talks to the RaceManager.
func (b *CozyBot) handleHostSubcommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	subCommand := options[0].Name

	var response string

	switch subCommand {
	case "open":
		// Call the RaceManager (Assuming b.Manager is accessible)
		// b.Manager.Phase = PhaseIdle
		response = "🟢 **The Green Room is OPEN.** Queue is accepting racers."

	case "start":
		// b.Manager.TriggerStaging()
		response = "🟠 **Ignition Sequence Initiated.** Staging in progress."

	case "bonus":
		amount := options[0].Options[0].IntValue()
		// b.Manager.AddToPot(amount)
		response = fmt.Sprintf("💸 **Sponsor Bonus!** Added $%d.00 to the pot.", amount/100)
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: response},
	})
}

// Helper to check for Admin permissions
func hasAdminPermissions(s *discordgo.Session, m *discordgo.Member) bool {
	// In reality, check for a specific Role ID here
	// For now, we just check if they have the "Administrator" permission bit
	return m.Permissions&discordgo.PermissionAdministrator != 0
}

The Full Function List Overview
Here is a summary table of the functions we have defined in the Bot Interface:
| Function Name | Trigger | Visibility | Purpose |
|---|---|---|---|
| RegisterCommands | On Boot | System | Uploads the command definitions to Discord servers. |
| interactionHandler | Any Interaction | System | The "Traffic Cop" that routes clicks/commands to logic. |
| handleRaceLink | /race | Ephemeral | Generates JWT + Magic Link for the Dashboard. |
| handleStats | /stats | Public | Displays a user's simulated Firestore statistics. |
| handleLeaderboard | /leaderboard | Public | Shows top 5 earners (Growth driver). |
| handleHostSubcommand | /host ... | Admin Only | Controls the state of the Race Manager (Open/Start). |
| AnnounceRaceResults | Event Driven | Public | Not a command. Called by Backend when race ends. |
Would you like me to visualize the Sequence of the /host start command to see how the Admin's Discord command triggers the Android Phones to wake up?
