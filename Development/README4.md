This project is now a fully functional "Esports League in a Box." You have the game (Mario Kart), the organization (Discord), the broadcast (Twitch), the money (Commissions), and the glory (Trophies).
The final piece of the puzzle to make this look like a professional TV broadcast is The OBS Overlay (Stream HUD).
Right now, your website shows the data to people viewing the website. But people watching directly on Twitch just see your raw phone screen. We need to get that "Queue List" and "Sponsor Logo" directly onto the video feed.
I would build a dedicated "Browser Source" Overlay next.
1. The Concept: "The HUD"
We will create a special, hidden page on your website (e.g., /overlay) designed specifically for OBS (Open Broadcaster Software).
 * Transparent Background: So it sits on top of your game.
 * Auto-Updating: It polls your Go backend every 3 seconds.
 * Dynamic: It cycles through Sponsor logos and shows the "Next Up" racer list automatically.
2. The Go Handler (internal/web/server.go)
We add a lightweight handler that serves raw JSON data for the overlay to consume, plus the HTML page itself.
// 1. serveOverlay: The HTML page for OBS
func serveOverlay(w http.ResponseWriter, r *http.Request, fs embed.FS) {
    tmpl, _ := template.ParseFS(fs, "web/templates/overlay.html")
    tmpl.Execute(w, nil)
}

// 2. apiOverlayData: The JSON endpoint the overlay polls
func apiOverlayData(w http.ResponseWriter, r *http.Request, db *store.Store) {
    // Fetch Next 3 Racers
    queue, _ := db.GetNextBatch(3) 
    
    // Fetch Current Pot Size (from the monetization step)
    pot, _ := db.GetCurrentPot()

    response := map[string]interface{}{
        "queue": queue,
        "pot":   pot,
        "sponsor": "YourSponsor.png", // Logic to rotate this later
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

3. The Transparent HTML (web/templates/overlay.html)
This creates a vertical sidebar that you can place on the left or right of your stream.
<!DOCTYPE html>
<html>
<head>
    <style>
        /* Transparent Body */
        body { background: transparent; margin: 0; overflow: hidden; font-family: 'Segoe UI', sans-serif; }
        
        .hud-container {
            width: 300px;
            position: absolute;
            top: 50px;
            right: 20px;
            display: flex;
            flex-direction: column;
            gap: 15px;
        }

        /* The "Next Up" Card */
        .queue-card {
            background: rgba(0, 0, 0, 0.85); /* Semi-transparent black */
            border-left: 5px solid #ffae00;
            padding: 15px;
            border-radius: 0 10px 10px 0;
            color: white;
            box-shadow: 0 4px 6px rgba(0,0,0,0.3);
            transform: translateX(100%); /* Start hidden */
            animation: slideIn 0.5s forwards;
        }

        /* The Sponsor Pill */
        .sponsor-pill {
            background: white;
            padding: 10px;
            border-radius: 50px;
            text-align: center;
        }
        .sponsor-pill img { height: 40px; }

        @keyframes slideIn { to { transform: translateX(0); } }
        
        .racer-row { display: flex; align-items: center; margin-top: 5px; }
        .racer-name { font-weight: bold; margin-left: 10px; font-size: 1.2em;}
    </style>
</head>
<body>
    <div class="hud-container" id="hud">
        </div>

    <script>
        async function updateHUD() {
            const res = await fetch('/api/overlay-data');
            const data = await res.json();
            
            const container = document.getElementById('hud');
            
            // Generate HTML
            let html = `
                <div class="queue-card">
                    <div style="font-size:0.8em; color:#bbb;">NEXT UP</div>
                    ${data.queue.map(p => `
                        <div class="racer-row">
                            <div class="racer-name">${p.MKTUsername}</div>
                        </div>
                    `).join('')}
                </div>
                
                <div class="queue-card" style="border-color: #4caf50;">
                    <div style="font-size:0.8em; color:#bbb;">CURRENT PRIZE POT</div>
                    <div style="font-size:1.5em; font-weight:800; color:#4caf50;">$${data.pot}</div>
                </div>

                <div class="sponsor-pill">
                    <img src="/static/sponsors/${data.sponsor}">
                </div>
            `;
            
            container.innerHTML = html;
        }

        // Poll every 3 seconds
        setInterval(updateHUD, 3000);
        updateHUD();
    </script>
</body>
</html>

4. How to Connect it to OBS
This is the "Magic Trick" that makes your stream look professional.
 * Open OBS Studio.
 * Click + in the Sources box.
 * Select Browser.
 * URL: http://localhost:8080/overlay (or your cloud URL).
 * Width: 1920, Height: 1080.
 * Result: The queue and prize pot will magically float over your gameplay video.
5. Diagram: The Broadcast Pipeline
Here is how the data flows from the database all the way to the Twitch viewer's screen.
This completes the visual identity of Cozy-Kart. Now, even if someone is watching on a smart TV (where they can't see your interactive website), they still see the Queue, the Pot, and the Sponsors directly on the video feed.
