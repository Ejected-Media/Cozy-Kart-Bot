This is a brilliant business strategy. You are essentially turning every race lobby into a strictly timed "Pop-Up Shop" where the 8 racers become your affiliate marketing team.
Because they only get paid if a sale happens while they are driving, they have a massive incentive to tell their own followers: "Come watch me race right now and buy a t-shirt! I get a cut!"
Here is how we implement this "Live Revenue Share" system technically.
1. The Concept: "The Commission Window"
We treat every Lobby (Race) as a financial session.
 * Trigger: When the Host types /start, the "Commission Window" opens.
 * Action: If a spectator buys merchandise (via Stripe, Ko-Fi, or Shopify) during this window, the system flags that sale.
 * Reward: A percentage (e.g., 20%) is calculated and split evenly among the 8 active participants.
 * Close: When the race ends, the window closes.
2. The Database Updates (store/store.go)
We need two new tables: one to track money owed to players (ledger) and one to record incoming sales (sales).
-- Track the money each player has earned
CREATE TABLE balances (
    discord_id TEXT PRIMARY KEY,
    pending_balance REAL DEFAULT 0.0,  -- Money earned but not paid out
    total_paid REAL DEFAULT 0.0,       -- Lifetime earnings
    FOREIGN KEY(discord_id) REFERENCES players(discord_id)
);

-- Record every sale and link it to the specific race lobby
CREATE TABLE sales (
    id TEXT PRIMARY KEY,               -- Transaction ID from Stripe/Shopify
    lobby_id INTEGER,                  -- Which race was active?
    amount REAL,                       -- Total sale price (e.g., $25.00)
    commission_pool REAL,              -- The cut for players (e.g., $5.00)
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

3. The "Webhook" Listener (internal/web/webhook.go)
You need a way for your shop to tell your Go app, "Hey, someone just bought a hoodie!" We do this with a Webhook.
Here is a function that listens for a sale, checks which lobby is active, and distributes the cash.
func HandleStripeWebhook(w http.ResponseWriter, r *http.Request, db *store.Store) {
    // 1. Parse the incoming JSON from Stripe/Shopify
    var payload struct {
        Amount float64 `json:"amount"` // e.g., 25.00
        TxID   string  `json:"id"`
    }
    json.NewDecoder(r.Body).Decode(&payload)

    // 2. Find the CURRENTLY active lobby
    // (We assume there is only one active race at a time for the MVP)
    activeLobbyID, racers, err := db.GetActiveRaceParticipants()
    if err != nil {
        log.Println("No active race, commission goes to house.")
        return
    }

    // 3. Calculate the Split
    // Let's say 20% goes to the players
    commissionRate := 0.20
    totalPool := payload.Amount * commissionRate // $25 * 0.20 = $5.00
    
    playerShare := totalPool / float64(len(racers)) // $5.00 / 8 players = $0.62 each

    // 4. Update the Database (Distribute the money)
    for _, racer := range racers {
        db.Exec(`
            INSERT INTO balances (discord_id, pending_balance) 
            VALUES (?, ?) 
            ON CONFLICT(discord_id) 
            DO UPDATE SET pending_balance = pending_balance + ?`, 
            racer.DiscordID, playerShare, playerShare)
    }

    // 5. Save the Sale Record
    db.Exec("INSERT INTO sales (id, lobby_id, amount, commission_pool) VALUES (?, ?, ?, ?)", 
        payload.TxID, activeLobbyID, payload.Amount, totalPool)

    log.Printf("💰 Sale processed! Each racer earned $%.2f", playerShare)
}

4. The "Hype Widget" on the Dashboard
To maximize sales, you shouldn't hide this logic. You should display it prominently on the screen to guilt/encourage viewers to support the racers.
Visual UI Element: A "Live Pot" ticker overlay.
<div class="revenue-ticker">
    <div class="label">RACER PRIZE POOL</div>
    <div class="amount" id="pool-amount">$0.00</div>
    <div class="subtext">Buy a shirt now to increase the pot!</div>
</div>

<style>
    .revenue-ticker {
        position: absolute;
        top: 20px;
        right: 20px;
        background: linear-gradient(135deg, #ffd700, #ffae00); /* Gold */
        color: #000;
        padding: 15px;
        border-radius: 10px;
        box-shadow: 0 0 20px rgba(255, 215, 0, 0.5);
        text-align: center;
        animation: pulse 2s infinite;
    }
    .amount { font-size: 2em; font-weight: 800; font-family: monospace; }
    
    @keyframes pulse {
        0% { transform: scale(1); }
        50% { transform: scale(1.05); }
        100% { transform: scale(1); }
    }
</style>

5. The Business Flow Diagram
Here is how the money moves from the Fan to the Player.
Critical Note on Payouts
Technically, calculating the money is easy (as shown above). Paying it out is the hard part due to tax laws (KYC, 1099 forms, etc.).
Recommended Start:
Instead of promising immediate "Cash," start by rewarding "Store Credit" or "Gift Cards."
 * Example: "If you earn $50 in commissions, we send you a $50 Amazon Gift Card or a Free Hoodie."
 * This avoids you having to become a regulated payment processor immediately.
Would you like me to write the Go code for the /wallet command so players can check their earnings in Discord?


___


This is a crucial feature for engagement. If players can see their balance going up, they will race more often and promote your stream harder.
Here is the implementation for the /wallet command.
1. The Database Logic (store/store.go)
Add this method to your Store struct. It fetches the balance, returning 0.00 if they are new.
// GetBalance returns the pending and paid amounts for a user
func (s *Store) GetBalance(discordID string) (float64, float64, error) {
    var pending, paid float64
    
    // We use COALESCE to return 0 instead of NULL if the row doesn't exist
    query := `
    SELECT COALESCE(pending_balance, 0), COALESCE(total_paid, 0) 
    FROM balances 
    WHERE discord_id = ?
    `
    err := s.DB.QueryRow(query, discordID).Scan(&pending, &paid)
    
    // If no row found, it's not an error, they just have $0
    if err == sql.ErrNoRows {
        return 0.0, 0.0, nil
    }
    
    return pending, paid, err
}

2. The Command Handler (internal/bot/bot.go)
We will add a new case to your interaction switch for "wallet".
Critical Detail: We set Flags: discordgo.MessageFlagsEphemeral. This ensures the message is private. Only the user who typed the command can see their money; it won't show up in the public chat.
func handleWallet(s *discordgo.Session, i *discordgo.InteractionCreate, db *store.Store) {
    user := i.Member.User
    
    // 1. Get the money
    pending, paid, err := db.GetBalance(user.ID)
    if err != nil {
        sendError(s, i, "Could not fetch wallet.")
        return
    }

    // 2. Format the response
    // We use an "Embed" for a nicer look with colors
    embed := &discordgo.MessageEmbed{
        Title:       "💰 Your Racer Wallet",
        Description: "Here is your share of the merch sales from races you participated in.",
        Color:       0xFFD700, // Gold color
        Thumbnail: &discordgo.MessageEmbedThumbnail{
            URL: user.AvatarURL(""),
        },
        Fields: []*discordgo.MessageEmbedField{
            {
                Name:   "Current Balance (Unpaid)",
                Value:  fmt.Sprintf("**$%.2f**", pending),
                Inline: true,
            },
            {
                Name:   "Lifetime Earnings",
                Value:  fmt.Sprintf("$%.2f", paid),
                Inline: true,
            },
        },
        Footer: &discordgo.MessageEmbedFooter{
            Text: "Payouts occur when balance reaches $20.00",
        },
    }

    // 3. Send Private Response
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Embeds: []*discordgo.MessageEmbed{embed},
            Flags:  discordgo.MessageFlagsEphemeral, // Only YOU see this
        },
    })
}

// Helper to register the command in your Start() function:
/*
    {
        Name:        "wallet",
        Description: "Check your earnings from merch sales",
    },
*/

3. The Visual Flow
Here is how the user experiences this "banking" feature within Discord:
4. Next Step: The "Payout" Command (Admin Only)
Eventually, you will need to actually pay them. You can add a hidden admin command:
 * Command: /payout @User 50.00
 * Logic:
   * Reduces pending_balance by 50.
   * Increases total_paid by 50.
   * Sends a public congratulatory message: "💸 @User just cashed out $50.00!"
This completes the economic loop of Cozy-Kart: Race -> Sell -> Earn -> Cash Out.

