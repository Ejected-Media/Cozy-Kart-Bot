package web

import (
	"cozy-kart/internal/store"
	"embed"
	"html/template"
	"log"
	"net/http"
)

// We pass the filesystem in from main.go to keep this package clean
func StartServer(db *store.Store, fs embed.FS, twitchUser string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		grid, _ := db.GetActiveGrid()

		// Fill empty slots if less than 8 players (for visual grid)
		displayGrid := make([]store.Player, 8)
		for i := 0; i < 8; i++ {
			if i < len(grid) {
				displayGrid[i] = grid[i]
			} else {
				// Empty placeholder
				displayGrid[i] = store.Player{MKTUsername: "WAITING...", Status: "EMPTY"}
			}
		}

		data := struct {
			TwitchUser string
			Grid       []store.Player
		}{
			TwitchUser: twitchUser,
			Grid:       displayGrid,
		}

		// Parse template
		tmpl, err := template.ParseFS(fs, "web/templates/index.html")
		if err != nil {
			http.Error(w, "Template error: "+err.Error(), 500)
			return
		}
		tmpl.Execute(w, data)
	})

	log.Println("🌐 Web Dashboard running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
