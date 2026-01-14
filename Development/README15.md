Here is the source code for race_manager.go.
This file acts as the central nervous system for the Racer Control Panel. It manages the "Green Room" queue, handles the "Key Ignition" security check, and tells the dashboard when to switch from "Idle" to "Race Mode."
package manager

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"cozy-kart/payout" // Assumes the previous package is imported
)

// RacePhase tracks the lifecycle of a single match.
type RacePhase int

const (
	PhaseIdle     RacePhase = iota // Waiting for players
	PhaseStaging                   // "Green Room" - Queue is locked, waiting for Ready Up
	PhaseIgnition                  // "Key Ignition" - Lobby code revealed
	PhaseRacing                    // "Compact Mode" - Race in progress
	PhaseCooldown                  // "After-Action" - Stats and Payouts
)

// Config holds the settings for a specific league event.
type Config struct {
	MaxRacers      int
	ReadyTimeout   time.Duration // How long they have to hit the big button
	LobbyCode      string        // The secret Mario Kart room code
}

// Racer acts as the connection between the Manager and the Payout session.
type Racer struct {
	ID        string
	Name      string
	IsReady   bool
	Session   *payout.RacerSession
	JoinTime  time.Time
}

// RaceManager orchestrates the flow of the event.
type RaceManager struct {
	Phase       RacePhase
	CurrentRaceID string
	Config      Config
	
	// The Queue
	Queue       []*Racer
	ActiveRacers map[string]*Racer

	// Concurrency
	mu sync.RWMutex
	
	// Notification channel (Simulates WebSocket push to UI)
	UpdateChan chan string
}

// NewRaceManager initializes the lobby controller.
func NewRaceManager(cfg Config) *RaceManager {
	return &RaceManager{
		Phase:        PhaseIdle,
		Config:       cfg,
		Queue:        make([]*Racer, 0),
		ActiveRacers: make(map[string]*Racer),
		UpdateChan:   make(chan string, 10),
	}
}

// EnqueueRacer adds a player to the "Green Room" waitlist.
func (rm *RaceManager) EnqueueRacer(id, name string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.Phase != PhaseIdle {
		return errors.New("cannot join queue: race is already in staging or progress")
	}

	newRacer := &Racer{
		ID:       id,
		Name:     name,
		IsReady:  false,
		Session:  payout.NewRacerSession(id), // Link to their bank
		JoinTime: time.Now(),
	}

	rm.Queue = append(rm.Queue, newRacer)
	rm.ActiveRacers[id] = newRacer
	
	rm.broadcast(fmt.Sprintf("USER_JOIN: %s joined the Green Room (Pos: #%d)", name, len(rm.Queue)))
	return nil
}

// TriggerStaging moves the lobby to the "Ready Check" phase.
// This lights up the "Ready Up" button on the UI.
func (rm *RaceManager) TriggerStaging() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if len(rm.Queue) < 2 {
		rm.broadcast("ERROR: Not enough racers to start staging.")
		return
	}

	rm.Phase = PhaseStaging
	rm.broadcast("STATUS_CHANGE: STAGING. Please click READY!")
	
	// Start a timer in a goroutine to disqualify AFK players
	go rm.monitorReadyStatus()
}

// ConfirmReady is called when the user clicks the Big Button in the Green Room.
func (rm *RaceManager) ConfirmReady(racerID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.Phase != PhaseStaging {
		return errors.New("ready check not active")
	}

	racer, exists := rm.ActiveRacers[racerID]
	if !exists {
		return errors.New("racer not found")
	}

	racer.IsReady = true
	rm.broadcast(fmt.Sprintf("READY_CONFIRM: %s is locked in.", racer.Name))

	// Check if everyone is ready
	allReady := true
	for _, r := range rm.Queue {
		if !r.IsReady {
			allReady = false
			break
		}
	}

	if allReady {
		// If everyone is ready, we unlock the "Ignition" phase immediately
		go rm.TriggerIgnition() 
	}

	return nil
}

// TriggerIgnition reveals the secure Lobby Code to the dashboard.
// This is the "Key Ignition" UI state.
func (rm *RaceManager) TriggerIgnition() {
	rm.mu.Lock()
	rm.Phase = PhaseIgnition
	code := rm.Config.LobbyCode
	rm.mu.Unlock()

	// Send the secret code ONLY to the active dashboard clients (simulated here)
	rm.broadcast(fmt.Sprintf("IGNITION_START: Secure Code Revealed: [%s]. Start your engines.", code))
	
	// Simulate time to type code before race actually starts
	time.Sleep(10 * time.Second)
	rm.StartRace()
}

// StartRace switches the UI to "Compact Mode" and enables payout accumulation.
func (rm *RaceManager) StartRace() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.Phase = PhaseRacing
	rm.CurrentRaceID = fmt.Sprintf("race_%d", time.Now().Unix())
	rm.broadcast("RACE_START: Dashboard minimizing to Compact Mode. Good luck!")
}

// EndRace finalizes the event and triggers payouts.
func (rm *RaceManager) EndRace() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.Phase = PhaseCooldown
	rm.broadcast("RACE_END: Calculating results...")

	for _, racer := range rm.Queue {
		// In a real implementation, we'd check if they actually finished via game telemetry.
		// Here we assume everyone finished for the sake of the example.
		payoutAmt, err := racer.Session.FinalizeRace(true)
		if err != nil {
			fmt.Printf("Error paying out %s: %v\n", racer.Name, err)
		} else {
			rm.broadcast(fmt.Sprintf("PAYOUT: %s earned %s", racer.Name, payoutAmt))
		}
	}
	
	// Reset for next race
	rm.Queue = make([]*Racer, 0)
	rm.ActiveRacers = make(map[string]*Racer)
	rm.Phase = PhaseIdle
	rm.broadcast("STATUS_CHANGE: IDLE. Waiting for next group.")
}

// monitorReadyStatus enforces the timeout for the "Ready Up" button.
func (rm *RaceManager) monitorReadyStatus() {
	time.Sleep(rm.Config.ReadyTimeout)
	
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.Phase != PhaseStaging {
		return
	}

	// Purge AFK players
	activeQueue := make([]*Racer, 0)
	for _, racer := range rm.Queue {
		if racer.IsReady {
			activeQueue = append(activeQueue, racer)
		} else {
			rm.broadcast(fmt.Sprintf("DQ: %s failed to ready up.", racer.Name))
			delete(rm.ActiveRacers, racer.ID)
		}
	}
	rm.Queue = activeQueue
	
	// Note: In a real app, you'd handle the logic of "Do we still have enough players?" here.
}

// Helper to push updates to the mock WebSocket
func (rm *RaceManager) broadcast(msg string) {
	select {
	case rm.UpdateChan <- msg:
	default:
		// Channel full, drop message or handle gracefully
	}
	fmt.Println("[MANAGER]:", msg)
}

Key Integrations with your UI:
 * Phase Management: The Phase constants (PhaseStaging, PhaseIgnition) map directly to the UI states we visualized earlier (The Green Room -> Key Ignition -> Compact Mode).
 * Ignition Logic: The TriggerIgnition method securely holds the lobby code and only releases it when all players are confirmed Ready, preventing "stream snipers" from stealing a spot before the racers get in.
 * DQ System: The monitorReadyStatus goroutine handles the anxiety-inducing countdown on the "Ready Up" button. If the timer hits zero, the racer is kicked from the queue.
Would you like to see the main.go file that wires payout.go and race_manager.go together into a running CLI simulation?
