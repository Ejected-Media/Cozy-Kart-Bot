We haven't explicitly built the Web Client folder structure yet. This client serves two purposes:
 * The Stream Overlay: What the audience sees on Twitch/YouTube (OBS Browser Source).
 * The Admin Dashboard: A web control panel for you (optional, but easier than Discord commands for complex tasks).
Here is the file map for the web/ directory. I recommend using Vite + React (or Svelte) for this, as it is fast and handles the high-frequency state updates from the Race Manager well.
📂 cozy-kart/web/
web/
├── public/
│   ├── assets/              # Static images (logos, placeholder avatars)
│   └── sounds/              # SFX (Cash register cha-ching, engine revs)
│
├── src/
│   ├── api/
│   │   └── socket.js        # The WebSocket Client (Connects to localhost:8080/ws)
│   │
│   ├── assets/              # CSS and SVGs
│   │   ├── fonts/           # Digital/LCD fonts for the HUD
│   │   └── main.css         # Global variables (Neon colors)
│   │
│   ├── components/          # Reusable UI Widgets
│   │   ├── PotDisplay.jsx   # The massive animated money counter
│   │   ├── Timer.jsx        # Race clock
│   │   ├── Leaderboard.jsx  # Vertical list of top racers
│   │   └── ChatTicker.jsx   # Scrolling display of recent donations
│   │
│   ├── layouts/             # The Main Screens
│   │   ├── Overlay.jsx      # The transparency-enabled view for OBS
│   │   └── AdminPanel.jsx   # The control center for the Host
│   │
│   ├── hooks/
│   │   └── useRaceData.js   # React Hook that subscribes to socket.js
│   │
│   ├── App.jsx              # Router (Switch between Overlay and Admin)
│   └── main.jsx             # Entry point
│
├── package.json             # Dependencies (React, GSAP for animations)
└── vite.config.js           # Build configuration

Key Integrations with the Backend:
 * src/api/socket.js:
   This file is the mirror image of the Go websocket.go file we just wrote. It establishes the connection and keeps it alive.
   // Simple reconnecting websocket logic
export const connectToRace = (onStateChange) => {
    const ws = new WebSocket("ws://localhost:8080/ws");

    ws.onmessage = (event) => {
        const state = JSON.parse(event.data);
        onStateChange(state); // Update React State
    };

    ws.onclose = () => {
        // Auto-reconnect if server restarts
        setTimeout(() => connectToRace(onStateChange), 1000);
    };
};

 * src/layouts/Overlay.jsx:
   This is designed to have a transparent background. You paste the URL http://localhost:3000/overlay into OBS. It only renders the UI elements (Pot, Timer, Leaderboard) so the gameplay footage shows through underneath.
 * src/components/PotDisplay.jsx:
   Since money is the focus, this component usually uses a library like react-spring or framer-motion to "roll" the numbers up (like a gas pump) whenever the backend sends a new pot amount.
Would you like to see the React Code for that Overlay.jsx component to see how it renders the dashboard?
