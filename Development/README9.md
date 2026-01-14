This is a fantastic addition. It leans heavily into the "Hangout" vibe and solves a common problem in esports streams: "What song is this?"
By turning a Racer into the "Designated DJ," you add another layer of personality to the broadcast. It also creates a new revenue stream (paid song requests).
Here is how we integrate the "Live DJ Booth" into the Cozy-Kart Dashboard.
1. The Concept: "Pass the Aux"
Since you likely have one main video feed (the Host's view of the race), the audio usually comes from that feed. However, we can overlay Music Metadata (Song Title, Artist, Album Art) synchronized with what the Host/DJ is playing.
 * The Feature: A "DJ Booth" widget replaces the "Cassette Player" when a race starts.
 * The Tech: We use the Last.fm API (it's free and easiest to work with). The Racer links their Last.fm/Spotify account, and we pull their "Now Scrobbling" track.
2. The Architecture: Fetching the Vibe
We need a new path for music data.
 * Racer plays music on Spotify/Apple Music while racing.
 * Last.fm Scrobbler (a standard app they install) sends this info to the cloud.
 * Cozy-Kart Backend polls the Last.fm API: "What is User X listening to?"
 * Dashboard Widget updates with the Album Art and Song Name.
3. The UI Layout: "The DJ Booth"
We place this in the bottom-left corner (balancing the Notification Dock on the right). It should look distinct—maybe a "Neon Sign" aesthetic or a "Vinyl Record" animation.
The HTML/CSS Component (web/templates/components/dj_booth.html):
<div class="dj-booth">
    <div class="dj-header">
        <span class="live-indicator">🔴 LIVE DJ</span>
        <span class="dj-name" id="dj-username">@RacerName</span>
    </div>
    
    <div class="track-card">
        <img id="album-art" src="/static/default-record.png" class="vinyl-spin">
        
        <div class="track-info">
            <div id="song-title" class="song-title">Sandstorm</div>
            <div id="artist-name" class="artist-name">Darude</div>
        </div>
    </div>

    <button class="request-btn" onclick="openTipJar()">
        💸 Request Song ($2)
    </button>
</div>

<style>
    .dj-booth {
        position: fixed;
        bottom: 20px;
        left: 20px;
        width: 280px;
        background: rgba(0, 0, 0, 0.8);
        border: 2px solid #bd00ff; /* Neon Purple */
        border-radius: 15px;
        padding: 15px;
        box-shadow: 0 0 15px rgba(189, 0, 255, 0.4);
        color: white;
        font-family: 'Segoe UI', sans-serif;
    }

    .dj-header {
        display: flex;
        justify-content: space-between;
        margin-bottom: 10px;
        font-size: 0.8em;
        text-transform: uppercase;
        letter-spacing: 1px;
    }

    .live-indicator { color: #ff4444; font-weight: bold; animation: blink 2s infinite; }

    .track-card {
        display: flex;
        align-items: center;
        gap: 15px;
    }

    .vinyl-spin {
        width: 60px; height: 60px;
        border-radius: 50%;
        border: 2px solid #333;
        animation: spin 3s linear infinite;
    }

    .song-title { font-weight: bold; font-size: 1.1em; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 170px; }
    .artist-name { color: #aaa; font-size: 0.9em; }

    .request-btn {
        width: 100%;
        margin-top: 12px;
        background: linear-gradient(45deg, #bd00ff, #5865F2);
        border: none;
        padding: 8px;
        border-radius: 5px;
        color: white;
        font-weight: bold;
        cursor: pointer;
        transition: transform 0.2s;
    }
    .request-btn:hover { transform: scale(1.05); }

    @keyframes spin { 100% { transform: rotate(360deg); } }
    @keyframes blink { 50% { opacity: 0.5; } }
</style>

4. The Backend Logic (internal/web/music.go)
You will need a free API Key from Last.fm.
package web

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type TrackInfo struct {
	Title  string
	Artist string
	Image  string
	IsPlaying bool
}

func GetNowPlaying(lastFmUser, apiKey string) TrackInfo {
	url := fmt.Sprintf("http://ws.audioscrobbler.com/2.0/?method=user.getrecenttracks&user=%s&api_key=%s&format=json&limit=1", lastFmUser, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return TrackInfo{Title: "Lofi Beats", Artist: "Cozy Radio", IsPlaying: false}
	}
	defer resp.Body.Close()

	// Parse JSON (Simplified for brevity)
	var result struct {
		RecentTracks struct {
			Track []struct {
				Name   string `json:"name"`
				Artist struct { Text string `json:"#text"` } `json:"artist"`
				Image  []struct { Text string `json:"#text"` } `json:"image"`
				Attr   struct { NowPlaying string `json:"nowplaying"` } `json:"@attr"`
			} `json:"track"`
		} `json:"recenttracks"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if len(result.RecentTracks.Track) > 0 {
		t := result.RecentTracks.Track[0]
		// Check if it's actually playing right now
		isPlaying := t.Attr.NowPlaying == "true"
		
		return TrackInfo{
			Title:     t.Name,
			Artist:    t.Artist.Text,
			Image:     t.Image[2].Text, // Index 2 is usually "Large" size
			IsPlaying: isPlaying,
		}
	}

	return TrackInfo{Title: "Offline", Artist: "", IsPlaying: false}
}

5. Monetization: "Tip the DJ"
This is where the business model shines.
 * The Trigger: User clicks "Request Song ($2)" on the dashboard.
 * The Payment: Opens a Stripe/Ko-Fi link (e.g., ko-fi.com/YourLeague).
 * The Loop: The user puts the song name in the donation message.
 * The Output: Your Discord Bot sends a DM to the current DJ: "💰 New Tip! Request: 'Play Free Bird' by Skynyrd."
6. The Updated "Home Portal" Dashboard
Here is the final visual layout of your product, combining everything we have designed: the Portal Widgets, the Stream, and now the DJ Booth.
Summary
By adding the DJ Booth, you solve three problems at once:
 * Content: The stream is never silent; music is always playing.
 * Credit: Viewers instantly know the song name without spamming chat.
 * Revenue: Tipping for requests is a proven money-maker on Twitch.
Would you like me to create the "Request Song" form logic (using a simple HTML modal) so users can type their request directly on the site before paying?
