Based on the comment in your earlier code (// Replace with Firestore call later), we will implement this using Google Cloud Firestore.
However, to keep your code "Cozy" and testable, we will use the Repository Pattern. This means the Game Engine (RaceManager) never talks to the database directly. It talks to an Interface.
This allows you to swap Firestore for SQLite, a JSON file, or even just memory (for testing) without breaking the game logic.
1. The Interface (pkg/repository/interface.go)
This defines the contract. The Game Engine says: "I don't care how you save it, just save the racer's earnings."
package repository

import "context"

// RacerStats holds the long-term data we want to keep.
type RacerStats struct {
    ID             string  `firestore:"id"` // Discord User ID
    Username       string  `firestore:"username"`
    TotalEarnings  float64 `firestore:"total_earnings"`
    RacesRun       int     `firestore:"races_run"`
    Wins           int     `firestore:"wins"`
}

// Repository is the interface our Database implementation must satisfy.
type Repository interface {
    // GetRacer finds a driver or returns an empty profile if new
    GetRacer(ctx context.Context, racerID string) (*RacerStats, error)
    
    // UpdateEarnings adds the new winnings to the total
    UpdateEarnings(ctx context.Context, racerID string, amount float64, isWin bool) error
    
    // Close cleans up connections
    Close() error
}

2. The Implementation (pkg/repository/firestore.go)
Now we write the actual code that talks to Google.
package repository

import (
    "context"
    "log"

    "cloud.google.com/go/firestore"
    "google.golang.org/api/iterator"
)

type FirestoreRepo struct {
    client *firestore.Client
}

// NewFirestore creates the connection using the GOOGLE_APPLICATION_CREDENTIALS json file.
func NewFirestore(projectID string) (*FirestoreRepo, error) {
    ctx := context.Background()
    client, err := firestore.NewClient(ctx, projectID)
    if err != nil {
        return nil, err
    }
    return &FirestoreRepo{client: client}, nil
}

func (r *FirestoreRepo) GetRacer(ctx context.Context, racerID string) (*RacerStats, error) {
    dsnap, err := r.client.Collection("racers").Doc(racerID).Get(ctx)
    if err != nil {
        // If not found, return a fresh stats object (don't error out)
        if status.Code(err) == codes.NotFound {
            return &RacerStats{ID: racerID}, nil
        }
        return nil, err
    }

    var stats RacerStats
    if err := dsnap.DataTo(&stats); err != nil {
        return nil, err
    }
    return &stats, nil
}

func (r *FirestoreRepo) UpdateEarnings(ctx context.Context, racerID string, amount float64, isWin bool) error {
    // We use a Transaction to ensure atomic updates (no race conditions)
    return r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
        docRef := r.client.Collection("racers").Doc(racerID)
        
        doc, err := tx.Get(docRef)
        var currentStats RacerStats
        
        if err != nil {
             // If it doesn't exist, we start at zero
             currentStats = RacerStats{ID: racerID}
        } else {
             if err := doc.DataTo(&currentStats); err != nil {
                 return err
             }
        }

        // Apply the math
        currentStats.TotalEarnings += amount
        currentStats.RacesRun += 1
        if isWin {
            currentStats.Wins += 1
        }
        
        // Save it back
        return tx.Set(docRef, currentStats)
    })
}

func (r *FirestoreRepo) Close() error {
    return r.client.Close()
}

3. Wiring it into the Manager (pkg/game/manager.go)
We update the RaceManager struct to hold this repository. Notice how we inject the repo in the constructor.
package game

import (
    "context"
    "log"
    "github.com/yourname/cozy-kart/pkg/repository"
)

type RaceManager struct {
    // ... existing fields (Pot, Mutex, Clients) ...
    
    Repo repository.Repository // The Abstract DB
}

func NewRaceManager(repo repository.Repository) *RaceManager {
    return &RaceManager{
        Repo: repo,
        // ... init other fields ...
    }
}

// EndGame is called when the timer hits 0:00
func (m *RaceManager) EndGame() {
    m.mu.Lock()
    defer m.mu.Unlock()

    // 1. Calculate Payouts
    // (Logic to split pot among top 3...)
    
    // 2. Save to DB asynchronously (don't block the UI)
    go func() {
        ctx := context.Background()
        
        for _, racer := range m.Racers {
            // Assume we calculated their 'winnings' locally
            isWinner := racer.Position == 1
            
            err := m.Repo.UpdateEarnings(ctx, racer.ID, racer.SessionEarnings, isWinner)
            if err != nil {
                log.Printf("❌ Failed to save stats for %s: %v", racer.Name, err)
            }
        }
        log.Println("💾 Career stats updated in Firestore.")
    }()
    
    m.State = StateIdle
}

4. The Final Update to main.go
We need to initialize the DB before we initialize the Game.
func main() {
    // ... load env ...

    // 1. Init Database
    // We can swap this line for NewSQLiteRepo() later if we want!
    db, err := repository.NewFirestore(os.Getenv("GCP_PROJECT_ID"))
    if err != nil {
        log.Fatalf("❌ DB Init Failed: %v", err)
    }
    defer db.Close()

    // 2. Inject DB into Game Manager
    raceManager := game.NewRaceManager(db)

    // ... start Bot and API ...
}

Why this structure is safe:
 * Transactions: The UpdateEarnings function uses a Firestore Transaction. If two races finish at the exact same millisecond (unlikely, but possible in distributed systems), the database won't lose money. It locks the record, updates it, and unlocks it.
 * Async Saves: inside EndGame, we wrap the DB calls in go func(). This ensures the specific "Race Finished!" animation on the phones triggers instantly, without waiting for the Google Cloud API to respond.
This completes the Core Backend Stack:
 * Logic: Go (RaceManager)
 * Comms: WebSockets (Gorilla)
 * UI: HTMX + Alpine (Templates)
 * Storage: Firestore (Repository Pattern)
Would you like to wrap this up by creating a Dockerfile so you can deploy this entire stack to a server (like Render, Railway, or a Raspberry Pi)?
