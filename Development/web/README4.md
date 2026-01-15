Here is the implementation for the Leaderboard Fragment.
This demonstrates the true power of the GOHA stack (Go + HTMX + Alpine). Instead of sending a complex JSON array and forcing the browser to iterate through it with JavaScript, Go does the heavy lifting. It iterates over the racers, applies logic (like "Who is in 1st place?"), and ships the finished HTML.
1. The Data Structure (Go Side)
First, let's briefly look at what data the template receives. In pkg/game/types.go:
type Racer struct {
    Name           string
    AvatarURL      string // Discord PFP
    Earnings       float64
    FormattedMoney string // "$12.50" (Pre-formatted in Go)
    Position       int    // 1, 2, 3...
}

type GameState struct {
    // ... pot, timer ...
    Leaderboard []Racer // The list we will loop over
}

2. The Template (web/templates/fragments/leaderboard.html)
We use the Go {{ range }} action to loop.
Key Design Choice: We use a simple if check inside the loop to give the 1st Place racer a special "Gold" color and a larger font size.
{{ define "leaderboard.html" }}

<div id="leaderboard" hx-swap-oob="true" class="w-full max-w-md bg-gray-900/80 rounded-xl overflow-hidden border border-gray-700 backdrop-blur-sm">
    
    <div class="flex justify-between items-center px-4 py-2 bg-black/50 border-b border-gray-700">
        <span class="text-xs font-bold text-gray-500 uppercase tracking-widest">Driver</span>
        <span class="text-xs font-bold text-gray-500 uppercase tracking-widest">Earnings</span>
    </div>

    <div class="flex flex-col">
        {{ range .Leaderboard }}
        
        <div class="flex items-center justify-between px-4 py-3 border-b border-gray-800 last:border-0 relative overflow-hidden group">
            
            <div class="flex items-center gap-3 z-10">
                <span class="font-mono font-bold text-lg w-6 
                    {{ if eq .Position 1 }} text-yellow-400 {{ else }} text-gray-600 {{ end }}">
                    #{{ .Position }}
                </span>

                <img src="{{ .AvatarURL }}" class="w-8 h-8 rounded-full border border-gray-600">
                
                <span class="font-bold text-gray-200 group-hover:text-white transition-colors">
                    {{ .Name }}
                </span>
            </div>

            <div x-data="{ pulse: true }" 
                 x-init="setTimeout(() => pulse = false, 500)"
                 class="font-mono font-bold transition-all duration-300"
                 :class="pulse ? 'text-white scale-105' : 'text-cyan-400 scale-100'"
            >
                {{ .FormattedMoney }}
            </div>

            {{ if eq .Position 1 }}
            <div class="absolute inset-0 bg-yellow-500/10 pointer-events-none"></div>
            {{ end }}

        </div>
        {{ end }} </div>
</div>
{{ end }}

3. The Backend Logic (Sorting)
Before sending this to the template, the Go backend needs to ensure the list is actually in order.
In pkg/game/manager.go:
func (m *RaceManager) GetState() GameState {
    // 1. Lock state to read safely
    m.mu.Lock()
    defer m.mu.Unlock()

    // 2. Sort the racers by Earnings (High to Low)
    // We use Go's standard sort library with a custom function.
    sort.SliceStable(m.Racers, func(i, j int) bool {
        return m.Racers[i].Earnings > m.Racers[j].Earnings
    })

    // 3. Assign Positions (1, 2, 3...) after sorting
    for i := range m.Racers {
        m.Racers[i].Position = i + 1
    }

    // 4. Return the View Model
    return GameState{
        Pot:         m.Pot,
        Leaderboard: m.Racers, // Now sorted and ranked
    }
}

4. Visualizing the Rendering Pipeline
It helps to see how the data flows from the struct into the visual table.
Why this is better than React for this specific case:
If you used React, you would send:
[{"id": 1, "name": "Speedy", "earnings": 50}, {"id": 2...}]
Then the Browser has to:
 * Parse JSON.
 * Run a JavaScript .map() loop.
 * Create DOM nodes.
 * Calculate if id === 1 to add the Gold class.
With Go Templates, the browser receives:
<div class="text-yellow-400">#1</div>
The logic is done. The browser just paints pixels. For a fast-paced overlay running on a potentially weak streaming PC, this is highly efficient.
We have the Pot (Single Value) and the Leaderboard (List).
The last fragment we typically need for a dashboard is the Recent Events / Chat Ticker (A list that only adds items, rather than replacing them).
Would you like to see how to handle a Scrolling Ticker (like "User X donated $5!") using HTMX's hx-swap="afterbegin"?
