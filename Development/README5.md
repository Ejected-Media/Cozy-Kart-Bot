This creates a hybrid product: part Game Lobby, part Personal Smart Dashboard (like a "Magic Mirror").
Since you are already running a Go server, fetching Weather APIs and checking Email servers is incredibly fast and efficient. We can overlay this information around the edges of the screen while the race stream plays in the center.
Here is how we upgrade Cozy-Kart into a Home Portal.
1. Architecture Update: The "Widget" Layer
We need to add two new data sources to your internal/web/server.go:
 * Weather Service: Fetches JSON from OpenWeatherMap (Free Tier).
 * Notification Service: Checks an IMAP email account (like Gmail) and Discord unreads.
2. The Weather Module (internal/web/weather.go)
First, we need a helper to get the weather. Since you are in Phoenix, we can default to that or make it configurable.
package web

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type WeatherData struct {
	Temp     int
	City     string
	Condition string
	Icon     string
}

// GetWeather fetches data from OpenWeatherMap
// You need a free API Key from openweathermap.org
func GetWeather(city, apiKey string) WeatherData {
	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s&units=imperial&appid=%s", city, apiKey)
	
	resp, err := http.Get(url)
	if err != nil {
		return WeatherData{Temp: 0, City: "Error", Condition: "Offline"}
	}
	defer resp.Body.Close()

	var result struct {
		Main struct { Temp float64 } `json:"main"`
		Weather []struct { Main string; Icon string } `json:"weather"`
		Name string `json:"name"`
	}
	
	json.NewDecoder(resp.Body).Decode(&result)

	// Safe defaults if parse fails
	condition := "Clear"
	icon := "01d"
	if len(result.Weather) > 0 {
		condition = result.Weather[0].Main
		icon = result.Weather[0].Icon
	}

	return WeatherData{
		Temp:      int(result.Main.Temp),
		City:      result.Name,
		Condition: condition,
		Icon:      "http://openweathermap.org/img/wn/" + icon + "@2x.png",
	}
}

3. The Notification Module (Discord + Email)
We can reuse your existing Discord Bot session to count users in the server, and add a simple IMAP check for email.
Update internal/web/server.go:
type DashboardWidgets struct {
    Time       string
    Weather    WeatherData
    MailCount  int
    DiscordMsg int
}

// In your main handler:
func renderHome(w http.ResponseWriter, r *http.Request) {
    // ... existing queue logic ...

    // 1. Get Weather (Hardcoded Phoenix for now)
    wx := GetWeather("Phoenix", "YOUR_OPENWEATHER_API_KEY")

    // 2. Get Notifications (Mock logic for speed, or implement IMAP)
    // Real IMAP logic requires the "go-imap" library
    mailCount := 4 
    
    // 3. Combine into data struct
    data := struct {
        Grid    []store.Player
        Widgets DashboardWidgets
        // ...
    }{
        Grid: grid,
        Widgets: DashboardWidgets{
            Weather:    wx,
            MailCount:  mailCount,
            DiscordMsg: 12, // You can pull this from dg.State.Guilds
        },
    }
    
    // ... execute template ...
}

4. The "Home Portal" UI Layout
We will transform the layout from a simple page into a HUD (Heads Up Display).
 * Top Left: Clock & Date (Large).
 * Top Right: Weather (Temp & Icon).
 * Bottom Right: Notification Badges (Mail, Discord).
 * Center: The Twitch Stream & Race Grid.
Update web/templates/index.html:
<div class="widget-bar">
    <div class="clock-widget">
        <div id="time" class="time-big">12:00 PM</div>
        <div id="date" class="date-small">Wednesday, Jan 14</div>
    </div>
    
    <div class="weather-widget">
        <div class="temp">{{ .Widgets.Weather.Temp }}°F</div>
        <div class="city">{{ .Widgets.Weather.City }}</div>
        <img src="{{ .Widgets.Weather.Icon }}" class="weather-icon">
    </div>
</div>

<div class="main-stage">
    </div>

<div class="dock">
    <div class="app-icon">
        📧 <span class="badge">{{ .Widgets.MailCount }}</span>
    </div>
    <div class="app-icon" style="background: #5865F2;">
        💬 <span class="badge">{{ .Widgets.DiscordMsg }}</span>
    </div>
</div>

<style>
    /* Widget Bar */
    .widget-bar {
        display: flex;
        justify-content: space-between;
        padding: 20px;
        background: linear-gradient(to bottom, rgba(0,0,0,0.9), transparent);
    }
    
    .time-big { font-size: 3em; font-weight: 200; font-family: 'Segoe UI Light', sans-serif; }
    .date-small { color: #aaa; font-size: 1.2em; }
    
    .weather-widget { text-align: right; display: flex; align-items: center; gap: 10px; }
    .temp { font-size: 2.5em; font-weight: bold; color: #ffae00; }

    /* The Dock */
    .dock {
        position: fixed;
        bottom: 20px;
        right: 20px;
        display: flex;
        gap: 15px;
        background: rgba(255,255,255,0.1);
        padding: 10px 20px;
        border-radius: 50px;
        backdrop-filter: blur(10px);
    }

    .app-icon {
        width: 50px; height: 50px;
        background: #444;
        border-radius: 50%;
        display: flex; justify-content: center; align-items: center;
        font-size: 1.5em;
        position: relative;
        cursor: pointer;
        transition: transform 0.2s;
    }
    .app-icon:hover { transform: translateY(-5px); }

    .badge {
        position: absolute;
        top: -5px; right: -5px;
        background: #ff4444;
        color: white;
        font-size: 0.7em;
        width: 20px; height: 20px;
        border-radius: 50%;
        display: flex; justify-content: center; align-items: center;
        border: 2px solid #222;
    }
</style>

<script>
    function updateClock() {
        const now = new Date();
        document.getElementById('time').innerText = now.toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'});
        document.getElementById('date').innerText = now.toLocaleDateString([], {weekday: 'long', month: 'short', day: 'numeric'});
    }
    setInterval(updateClock, 1000);
    updateClock();
</script>

5. Why this is powerful
You have effectively built your own "Cozy OS".
Instead of just checking a website when a race is happening, this dashboard becomes useful 24/7.
 * Morning: Check the weather in Phoenix while drinking coffee.
 * Work Day: Glance at the Notification Dock to see if you have emails.
 * Race Time: The center screen lights up with the Twitch stream and the Lobby Grid.
Next Step
Would you like me to refine the Weather Module to include a "Sunset" countdown (since the dashboard might look cool in "Dark Mode" at night), or move on to Hardware suggestions (like running this on a Raspberry Pi attached to a TV)?
