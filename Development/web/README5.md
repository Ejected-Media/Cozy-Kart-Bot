The "Ticker" (or Kill Feed/Donation Log) requires a different strategy.
For the Pot and Timer, we used State Synchronization (replacing the whole value 2x/second).
For the Ticker, we use Event Streams. When a donation happens, we want to inject a new item at the top of the list, without re-rendering the old ones.
We will use the advanced HTMX OOB syntax: hx-swap-oob="afterbegin:#id". This tells HTMX: "Don't replace the list. Just shove this new item into the top of it."
1. The Container (web/templates/overlay.html)
This is just an empty placeholder where the events will live.
<div id="event-feed" class="absolute bottom-10 left-10 w-80 h-48 overflow-hidden flex flex-col gap-2">
    </div>

2. The Single Item Fragment (web/templates/fragments/event.html)
This is the crucial part. We are not sending the whole list. We are sending one div.
We combine HTMX (for insertion) with Alpine (for entrance and exit animations).
{{ define "event.html" }}

<div hx-swap-oob="afterbegin:#event-feed">
    
    <div x-data="{ show: false }"
         x-init="
            setTimeout(() => show = true, 50); 
            setTimeout(() => { show = false; setTimeout(() => $el.remove(), 500) }, 5000)
         "
         x-show="show"
         x-transition:enter="transition ease-out duration-300"
         x-transition:enter-start="opacity-0 -translate-x-10"
         x-transition:enter-end="opacity-100 translate-x-0"
         x-transition:leave="transition ease-in duration-500"
         x-transition:leave-start="opacity-100"
         x-transition:leave-end="opacity-0"
         
         class="bg-black/80 border-l-4 border-yellow-400 p-3 rounded shadow-lg backdrop-blur-md"
    >
        <div class="flex justify-between items-center">
            <span class="font-bold text-white">{{ .Username }}</span>
            <span class="font-bold text-yellow-400">+${{ .Amount }}</span>
        </div>
        <div class="text-xs text-gray-400 italic">
            {{ .Message }}
        </div>
    </div>

</div>
{{ end }}

3. The Backend Trigger (pkg/api/websocket.go)
Unlike the polling loop (which runs every 500ms), we want this to be instant. We create a dedicated function that main.go or RaceManager can call immediately when a donation arrives.
// BroadcastEvent sends a single "Toast" notification to all screens.
func (h *Handler) BroadcastEvent(username string, amount float64, message string) {
    
    // 1. Prepare the data
    data := struct {
        Username string
        Amount   string // formatted
        Message  string
    }{
        Username: username,
        Amount:   fmt.Sprintf("%.2f", amount),
        Message:  message,
    }

    // 2. Render the single fragment
    var msgBuffer bytes.Buffer
    tmpl := template.Must(template.ParseGlob("web/templates/fragments/event.html"))
    
    err := tmpl.ExecuteTemplate(&msgBuffer, "event.html", data)
    if err != nil {
        log.Println("Template Error:", err)
        return
    }

    // 3. Push to everyone immediately
    h.mu.Lock()
    defer h.mu.Unlock()
    
    for conn := range h.Clients {
        conn.WriteMessage(websocket.TextMessage, msgBuffer.Bytes())
    }
}

4. Why this "Self-Destruct" Pattern is useful
If you notice the Alpine logic:
setTimeout(() => $el.remove(), 5000)
This is critical for long-running overlays. If you just kept appending divs (afterbegin) for a 2-hour race, the browser DOM would eventually have 1,000 hidden elements, potentially slowing down the stream.
By having the elements delete themselves after 5 seconds, the DOM stays clean, usually containing only the 3-4 visible notifications at any given time.
Summary of the "GOHA" Overlay Stack
We have now built all three core components of a broadcast overlay using zero client-side JavaScript (logic-wise):
 * The HUD (Pot/Timer): High-frequency OOB Replacement (Sync).
 * The Leaderboard: Complex List Rendering (Server-side logic).
 * The Ticker: Event-driven Injection (Append + Self-Destruct).
This setup is extremely robust. If the browser crashes and you refresh, it instantly gets the latest state. If the server restarts, the socket.js (which we wrote earlier) reconnects, and the feed resumes.
Would you like to move on to the Data Persistence layer (how we save these earnings to the database so they persist across reboots), or are you happy with this In-Memory version for now?
