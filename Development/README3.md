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
