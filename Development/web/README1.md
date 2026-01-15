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

___

Based on your interest in GoLang and keeping things efficient (and likely for a project like C500 or selfielunge where you want to focus on layout/content rather than complex state management), you are almost certainly thinking of HTMX.
The combination is often called the GOHA Stack (Go, HTMX, Alpine) or simply the Hypermedia Stack.
 * Alpine.js: Handles the "micro-interactions" on the client (toggling menus, simple animations, dropdowns) without leaving the HTML.
 * HTMX: Handles the server communication. Instead of fetching JSON and rebuilding the DOM (like React), it fetches ready-to-render HTML from your Go server and swaps it into the page.
 * Tailwind CSS: (Usually) handles the styling.
Why we likely chose this for the other project:
 * Go-Centric: With React (Cozy-Kart), your Go server is just a JSON vending machine. With HTMX, your Go server is the star. You write Go Templates (html/template) that render the UI directly.
 * No Build Step: You don't need npm, webpack, or vite. You can just drop the <script> tags for Alpine and HTMX into your header and go.
 * "Classic" Feel: It aligns perfectly with your interest in Classic Tech (MacOS 7/8, etc.). It returns to the "Old Web" philosophy where the server is in charge, but with the smoothness of a modern SPA.
Example: The "Active Search" Pattern
Here is how Alpine and HTMX work together in a Go template. Note how there is no JavaScript written here, just attributes.
<div x-data="{ query: '' }"> <input 
        type="text" 
        name="search"
        placeholder="Search Users..."
        
        /* Alpine: Sync input to local state */
        x-model="query"
        
        /* HTMX: When key goes up, wait 500ms, then hit the server */
        hx-post="/search" 
        hx-trigger="keyup changed delay:500ms" 
        hx-target="#search-results" 
        hx-indicator="#loading"
    >

    <img id="loading" class="htmx-indicator" src="/spinner.gif"/>

</div>

<div id="search-results">
    </div>

Does that sound right? Alpine for the UI polish, HTMX for the data fetching?

