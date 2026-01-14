// In package bot

// AnnounceRaceResults creates a stylish "Polaroid" embed for the public channel.
func (b *CozyBot) AnnounceRaceResults(channelID string, winnerName string, potSize int64, duration time.Duration) error {
	
	// Format money (e.g., 1050 cents -> "$10.50")
	formattedPot := fmt.Sprintf("$%.2f", float64(potSize)/100.0)

	// We use an "Embed" to make it look like a distinct card
	embed := &discordgo.MessageEmbed{
		Title:       "📸 Race Result: The Cocoa Cup",
		Description: fmt.Sprintf("**%s** crossed the finish line first!", winnerName),
		Color:       0x00FF7F, // Spring Green (Success)
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://your-bucket-url.com/cozy-kart-trophy.png",
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "🏆 Winner's Purse",
				Value:  formattedPot,
				Inline: true,
			},
			{
				Name:   "⏱️ Race Duration",
				Value:  duration.Round(time.Second).String(),
				Inline: true,
			},
			{
				Name:   "✨ Vibe Check",
				Value:  "Immaculate",
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Race ID: #8841-COZY • Replay available in App",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	_, err := b.Session.ChannelMessageSendEmbed(channelID, embed)
	return err
}
