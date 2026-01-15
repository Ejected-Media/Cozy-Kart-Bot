Here is the source code for the pot.html fragment.
The Secret Sauce: "Out of Band" Swaps
In standard HTMX, a response usually replaces the element that triggered the request. But with WebSockets, we aren't "replying" to a specific click—we are pushing data unprompted.
To do this, we use the Out of Band (OOB) attribute (hx-swap-oob="true").
This tells HTMX: "I don't care where I came from. Just find the element with this ID in the DOM and replace it."
📄 web/templates/fragments/pot.html
{{ define "pot.html" }}
<div id="pot-display" hx-swap-oob="true" class="absolute top-10 right-10 text-right">
    
    <div class="text-sm text-cyan-400 font-bold tracking-widest uppercase mb-1">
        Current Pot
    </div>

    <div x-data="{ show: false }"
         x-init="setTimeout(() => show = true, 50)"
         class="text-6xl font-black text-yellow-400 transition-all duration-300 transform"
         :class="show ? 'scale-110 text-white' : 'scale-100 text-yellow-400'"
    >
        {{ .FormattedPot }} 
    </div>

    <div class="mt-2 inline-block bg-gray-900 rounded-full px-3 py-1 border border-gray-700">
        <span class="text-xs text-gray-400">Racer Split: </span>
        <span class="text-xs text-green-400 font-bold">{{ .SplitPercent }}%</span>
    </div>

</div>
{{ end }}

How the Go Server Sends This
You don't just send the Pot. In the BroadcastLoop inside websocket.go, you can smash multiple fragments together into one message. This allows you to update the Timer, the Pot, and the Leaderboard all at once in a single "frame."
// Inside your BroadcastLoop...

// 1. Create a buffer to hold the HTML
var msgBuffer bytes.Buffer

// 2. Render the Pot Fragment into the buffer
// (Go looks for "pot.html" and fills in .FormattedPot)
tmpl.ExecuteTemplate(&msgBuffer, "pot.html", state)

// 3. Render the Timer Fragment into the SAME buffer
// (Go appends this <div id="timer">...</div> right after the pot)
tmpl.ExecuteTemplate(&msgBuffer, "timer.html", state)

// 4. Send the whole blob as one text message
conn.WriteMessage(websocket.TextMessage, msgBuffer.Bytes())

Why this works for "Cozy-Kart":
 * Atomic Updates: The pot and the timer update in perfect sync because they arrive in the same WebSocket packet.
 * Animation Trigger: Because HTMX physically removes the old div and inserts a new one, Alpine.js treats it as a brand new element. The x-init fires again, causing the text to "pop" (scale up) every single time the money changes.
We have the Pot covered. Would you like to do the Leaderboard fragment next? That one is trickier because it involves a loop (Go range) inside the HTML.
