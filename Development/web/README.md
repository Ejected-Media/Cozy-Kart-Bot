Here is the complete README.md for the Cozy-Kart project. You can save this file in the root of your repository to keep the vision and architecture clear as you build.
🏎️ Cozy-Kart: The Interactive Racing Dashboard
Cozy-Kart is a real-time, multi-device racing management system. It connects a Discord Community (the audience/racers) to a Live Broadcast (the stream) via a Mobile Companion App (the dashboard).
It is designed to turn a passive stream into an interactive event where viewers become racers, managing their "winnings" and "stats" on their physical phones while watching the race on the main screen.
🏗️ System Architecture
The system operates on a Hub-and-Spoke model with four distinct components:
 * The Brain (Go Backend): The central source of truth. It manages the race state, calculating payouts, and broadcasting updates via WebSockets.
 * The Controller (Discord Bot): The admin interface. Hosts use slash commands (/host start) to control the game state.
 * The Dashboard (Android App): The player's telemetry unit. A "Green Room" for waiting and a high-contrast HUD for racing.
 * The Overlay (Web/HTMX): The broadcast visualizer. A lightweight, server-rendered page for OBS.
The Tech Stack
| Component | Technology | Reasoning |
|---|---|---|
| Backend | Go (Golang) | High concurrency, low latency, single binary deployment. |
| Database | Firestore | Real-time capable, scalable, serverless. |
| Mobile | Kotlin / Compose | Native performance, declarative UI, deep Android integration. |
| Web / Overlay | HTMX + Alpine.js | "GOHA Stack." Zero build steps, server-side logic, high performance. |
| Communication | WebSockets | Persistent, bi-directional streams (Gorilla/Websocket). |
| Deployment | Docker | Multi-stage build for a tiny Alpine Linux container. |
📂 Project Structure
cozy-kart/
├── cmd/
│   └── cozy-kart/
│       └── main.go           # Application Entry Point (Wires Bot + API + Game)
│
├── pkg/
│   ├── api/
│   │   ├── handler.go        # HTTP Routes
│   │   └── websocket.go      # The Socket Hub (Fan-Out Logic)
│   ├── bot/
│   │   ├── bot.go            # Discord Session Manager
│   │   └── commands.go       # Slash Command Definitions (/race, /host)
│   ├── game/
│   │   ├── manager.go        # Core Game Loop (Timer, Pot Logic)
│   │   └── types.go          # Data Structs (Racer, GameState)
│   └── repository/
│       ├── interface.go      # DB Interface (Repository Pattern)
│       └── firestore.go      # Google Cloud Implementation
│
├── web/
│   ├── static/               # CSS, Fonts, JS Vendors
│   └── templates/
│       ├── layout.html       # Base HTML
│       ├── overlay.html      # Main OBS Page
│       └── fragments/        # HTMX Partials (pot.html, timer.html, list.html)
│
├── android/                  # (Separate Android Studio Project)
│   ├── app/src/main/java/com/example/cozykart/
│   │   ├── ui/
│   │   │   ├── GreenRoom.kt  # Waiting Area Screen
│   │   │   └── RaceHUD.kt    # Telemetry Screen
│   │   └── network/
│   │       └── SocketMgr.kt  # OkHttp WebSocket Client
│
├── Dockerfile                # Multi-stage build definition
├── go.mod                    # Go Dependencies
└── README.md                 # This file

🚀 Getting Started
1. Prerequisites
 * Go 1.23+
 * Docker Desktop
 * Google Cloud Account (for Firestore)
 * Discord Developer Account (for Bot Token)
2. Environment Variables
Create a .env file in the root directory:
# Discord Configuration
DISCORD_TOKEN=your_bot_token_here
DISCORD_APP_ID=your_application_id
DISCORD_GUILD_ID=your_server_id

# Google Cloud Configuration
GCP_PROJECT_ID=your_project_id
GOOGLE_APPLICATION_CREDENTIALS=./service-account.json

# Server Configuration
PORT=8080

3. Running Locally (Development)
# Run the Go Backend
go run cmd/cozy-kart/main.go

4. Running with Docker (Production)
# Build the container
docker build -t cozy-kart .

# Run the container
docker run -p 8080:8080 --env-file .env cozy-kart

🎮 Key Features Breakdown
A. The "Green Room" (Android)
 * Purpose: Keeps the racers engaged before the event starts.
 * Behavior: Shows connection status, server MOTD, and a "Ready" indicator.
 * Trigger: Automatically switches to the Race HUD when the backend broadcasts STATE: RACING.
B. The Race HUD (Android)
 * Design: "Digital Dashboard" aesthetic. Massive font sizes for the Money Counter.
 * Vibe Check: The entire UI shifts color (Cyan → Gold) when the "Pot Bonus" is active.
 * OLED Friendly: Uses #121212 dark backgrounds to save battery on mobile devices.
C. The Overlay (OBS/Web)
 * No JavaScript Logic: The logic lives in Go Templates.
 * HTMX Swapping: Uses hx-swap-oob="true" to surgically update the Pot and Timer without refreshing the page.
 * Self-Destructing Events: Donation notifications appear, animate in with Alpine.js, and then remove themselves from the DOM to keep performance high.
🔮 Future Roadmap
 * [ ] Twitch Integration: Allow chat commands (!boost) to directly affect the pot.
 * [ ] Physical Hardware: Add support for a physical "Big Red Button" using GPIO on a Raspberry Pi to start the race.
 * [ ] Season Mode: Track stats across multiple weeks using the Firestore wins counter.
 * [ ] 
