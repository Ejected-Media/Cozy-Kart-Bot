This project is shaping up to be a self-sustaining esports league!
Since we have the Lobby (Start), the Dashboard (Watch), and the Wallet (Earn), the missing piece of the puzzle is the Result.
Right now, the system knows a race started, but it doesn't know who won. Without results, we can't have a "Champion," and we can't award performance bonuses (e.g., "Winner takes 50% of the pot").
I would build the "League & Leaderboard" System next. This turns one-off races into a Season where points (and money) accumulate over time.
1. The "Report" Command (/result)
Since Mario Kart Tour doesn't have an API to tell us who won automatically, players must self-report (honor system), or the Host can report for them.
The Workflow:
 * Race ends.
 * User types /result position:1 (I got 1st place).
 * Bot confirms and updates the database.
 * If they lie? The Host can use an /admin-fix command.
The Code (internal/bot/bot.go):
// Add this to your command handler
case "result":
    // 1. Get the position (1-8)
    pos := i.ApplicationCommandData().Options[0].IntValue()
    
    // 2. Update the Database
    // We update the 'participants' table for the active lobby
    err := db.RecordResult(i.Member.User.ID, int(pos))
    
    if err != nil {
        sendError(s, i, "Could not save result. Is a race active?")
        return
    }
    
    // 3. Calculate "League Points" (Example: 1st=15pts, 8th=1pt)
    points := calculatePoints(int(pos))
    db.AddLeaguePoints(i.Member.User.ID, points)

    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: fmt.Sprintf("🏆 recorded! You finished **#%d** and earned **%d pts**.", pos, points),
        },
    })

2. The "Season Standings" Page
We need a new page on your website (or a section below the video) to show who is the King of the Season.
New SQL Table:
CREATE TABLE league_standings (
    discord_id TEXT PRIMARY KEY,
    points INTEGER DEFAULT 0,
    races_played INTEGER DEFAULT 0,
    wins INTEGER DEFAULT 0
);

The HTML Snippet (standings.html):
<div class="leaderboard-panel">
    <h3>🏆 Season 1 Standings</h3>
    <table>
        <tr>
            <th>Rank</th>
            <th>Racer</th>
            <th>Points</th>
            <th>Wins</th>
        </tr>
        {{ range .Leaderboard }}
        <tr>
            <td>#{{ .Rank }}</td>
            <td>
                <img src="{{ .AvatarURL }}" class="tiny-avatar">
                {{ .MKTUsername }}
            </td>
            <td><strong>{{ .Points }}</strong></td>
            <td>{{ .Wins }}</td>
        </tr>
        {{ end }}
    </table>
</div>

<style>
    .leaderboard-panel {
        background: #222;
        border: 2px solid #5865F2;
        border-radius: 10px;
        padding: 20px;
        margin-top: 30px;
    }
    table { width: 100%; border-collapse: collapse; }
    td, th { padding: 10px; text-align: left; border-bottom: 1px solid #444; }
    .tiny-avatar { width: 24px; height: 24px; border-radius: 50%; vertical-align: middle; margin-right: 8px;}
</style>

3. Visualizing the "Season Lifecycle"
Here is how the League Ecosystem feeds into itself. This diagram explains how the "Wallet" we built earlier interacts with the "Leaderboard" we just designed.
4. Why this matters for Business
This allows you to sell "Season Sponsorships".
Instead of selling ads for just one race, you can approach a sponsor and say:
> "Sponsor the 'Cozy-Kart Spring Season'. Your logo will be on the Leaderboard page for 3 months, and you will present the Grand Prize check to the winner."
> 
This is how real sports monetize, and it works perfectly for your community league.
Would you like me to write the calculatePoints function (standard Mario Kart scoring: 15, 12, 10, 8...) or help you design the "Trophy Case" UI?

___


It is exciting to see this evolve into a full League! Since you said "Yes," I will provide both the scoring logic and the visual Trophy Case design so you have the complete package.
1. The Scoring Logic (internal/bot/logic.go)
We will use a standard "Formula 1" or "Mario Kart" style point distribution. This rewards consistency but heavily favors winning.
package bot

// calculatePoints maps a finishing position (1-8) to a point value.
// Based on standard Mario Kart 8 Deluxe distribution.
func calculatePoints(position int) int {
	switch position {
	case 1:
		return 15 // 🥇 Gold
	case 2:
		return 12 // 🥈 Silver
	case 3:
		return 10 // 🥉 Bronze
	case 4:
		return 9
	case 5:
		return 8
	case 6:
		return 7
	case 7:
		return 6
	case 8:
		return 5
	default:
		return 1 // Participation point for DNFs or crashes
	}
}

// GetTrophyIcon returns a visual emoji for the top 3
func GetTrophyIcon(rank int) string {
	switch rank {
	case 1:
		return "🏆"
	case 2:
		return "🥈"
	case 3:
		return "🥉"
	default:
		return ""
	}
}

2. The "Trophy Case" UI (web/templates/profile.html)
A leaderboard is just a list of names. A Trophy Case is personal. This is a section you can add to the User's Profile page on your dashboard. It shows off their past victories visually.
The HTML:
<div class="trophy-case-container">
    <h2>🏆 Hall of Fame: {{ .Username }}</h2>
    
    <div class="stats-row">
        <div class="stat-box">
            <div class="value">{{ .SeasonPoints }}</div>
            <div class="label">Season Points</div>
        </div>
        <div class="stat-box">
            <div class="value">{{ .TotalWins }}</div>
            <div class="label">1st Place Wins</div>
        </div>
        <div class="stat-box money">
            <div class="value">${{ .Earnings }}</div>
            <div class="label">Career Earnings</div>
        </div>
    </div>

    <h3>Cabinet</h3>
    <div class="cabinet-grid">
        {{ range .Trophies }}
        <div class="trophy-item tooltip">
            <span class="icon">{{ .Icon }}</span> <span class="date">{{ .Date }}</span>
            <span class="tooltiptext">{{ .RaceName }}</span>
        </div>
        {{ else }}
        <div class="empty-case">No trophies yet... get racing!</div>
        {{ end }}
    </div>
</div>

The CSS:
.trophy-case-container {
    background: #1a1a1a;
    border-radius: 12px;
    padding: 20px;
    color: white;
    max-width: 500px;
    margin: 20px auto;
    border: 1px solid #333;
}

.stats-row {
    display: flex;
    justify-content: space-between;
    margin-bottom: 20px;
}

.stat-box {
    text-align: center;
    background: #2b2b2b;
    padding: 10px;
    border-radius: 8px;
    flex: 1;
    margin: 0 5px;
}

.stat-box.money .value { color: #4caf50; } /* Green for money */

.value { font-size: 1.5em; font-weight: bold; }
.label { font-size: 0.8em; color: #888; text-transform: uppercase; }

.cabinet-grid {
    display: grid;
    grid-template-columns: repeat(5, 1fr); /* 5 Trophies per row */
    gap: 10px;
    background: #000;
    padding: 15px;
    border-radius: 8px;
    border-top: 4px solid #d4af37; /* Gold Trim */
}

.trophy-item {
    font-size: 2em;
    text-align: center;
    cursor: pointer;
    transition: transform 0.2s;
}

.trophy-item:hover { transform: scale(1.2); }

.empty-case { grid-column: span 5; text-align: center; color: #555; padding: 20px; font-style: italic; }

3. The Visual Summary
Here is how the League System fits into your overall architecture. This diagram shows the complete loop from the moment a race finishes to the moment the trophy appears on the website.
Summary of What We Built
 * Engine: calculatePoints handles the math.
 * Display: The HTML/CSS creates a beautiful, retro-gaming style profile.
 * Incentive: Players now race not just for fun, but to fill their empty cabinet slots.
This effectively completes the "Cozy-Kart" ecosystem! You have Operations (Bot), Display (Web), Monetization (Commissions), and Retention (League/Trophies).
