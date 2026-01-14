package main

import (
	"fmt"
	"math/rand"
	"time"

	// In a real project, these would be:
	// "github.com/yourname/cozy-kart/manager"
	// "github.com/yourname/cozy-kart/payout"
	
	"cozy-kart/manager"
	"cozy-kart/payout"
)

func main() {
	// 1. Setup the League Configuration
	cfg := manager.Config{
		MaxRacers:    4,
		ReadyTimeout: 5 * time.Second,
		LobbyCode:    "8841-COZY", // The secret code
	}

	// 2. Initialize the Manager
	rm := manager.NewRaceManager(cfg)
	
	fmt.Println("=========================================")
	fmt.Println("   ☕ COZY-KART LEAGUE: SERVER ONLINE ☕   ")
	fmt.Println("=========================================")
	fmt.Println("[SYSTEM] Waiting for racers...")

	// 3. Simulate Racers Joining (The "Green Room")
	// In reality, this would be an HTTP handler or WebSocket event
	racers := []struct{ ID, Name string }{
		{"r1", "SpeedyBoi"},
		{"r2", "DriftQueen"},
		{"r3", "VibeCheck"},
	}

	for _, r := range racers {
		time.Sleep(500 * time.Millisecond)
		if err := rm.EnqueueRacer(r.ID, r.Name); err != nil {
			fmt.Printf("Error joining: %v\n", err)
		}
	}

	// 4. Trigger the "Ready Check" (Staging Phase)
	fmt.Println("\n[ADMIN] Triggers Ready Check...")
	rm.TriggerStaging()

	// 5. Simulate Racers clicking the [READY] button
	// We'll simulate one user being slow/AFK for drama, but eventually readying up
	go func() {
		time.Sleep(1 * time.Second)
		rm.ConfirmReady("r1")
		
		time.Sleep(500 * time.Millisecond)
		rm.ConfirmReady("r2")

		time.Sleep(1 * time.Second) // "VibeCheck" is taking their time sipping coffee
		rm.ConfirmReady("r3") 
	}()

	// Wait for the Ignition sequence (handled by the manager automatically once everyone is ready)
	// We monitor the channel for status updates
	raceStarted := false
	for !raceStarted {
		msg := <-rm.UpdateChan
		fmt.Println("[UI UPDATE]:", msg)
		
		// Crude check to see if race started so we can move to the next simulation step
		if len(msg) > 10 && msg[:10] == "RACE_START" {
			raceStarted = true
		}
	}

	// 6. THE RACE IS ON! (Simulate Chat & Payouts)
	// This simulates the "Paycheck" module updating in real-time
	fmt.Println("\n--- 🏁 RACE IN PROGRESS 🏁 ---")
	
	// We'll use a WaitGroup to simulate the race duration
	raceDuration := 4 * time.Second
	
	// Simulate Chat Donations (The "Pot")
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		
		for i := 0; i < 6; i++ { // 6 donation ticks
			<-ticker.C
			donation := payout.Money(rand.Int63n(1000) + 100) // Random donation $1.00 - $11.00
			
			// Apply donation to EVERY active racer's pot
			// (In a real app, we'd iterate safely, here we just hack it for the demo)
			for _, r := range racers {
				if activeRacer, ok := rm.ActiveRacers[r.ID]; ok {
					activeRacer.Session.UpdateLivePot(donation)
				}
			}
			fmt.Printf("💰 [DONATION] +%s added to the pot!\n", donation)
		}
	}()

	// Simulate "Vibe Check" (Bonus Split)
	time.Sleep(2 * time.Second)
	fmt.Println("✨ [VIBES] Chat is hyping up! Bonus Split Activated! ✨")
	for _, r := range racers {
		if activeRacer, ok := rm.ActiveRacers[r.ID]; ok {
			activeRacer.Session.ToggleVibeBonus(true)
		}
	}

	time.Sleep(raceDuration)

	// 7. Race Ends
	fmt.Println("\n--- 🏁 FINAL LAP 🏁 ---")
	rm.EndRace()
	
	// Drain the remaining messages from the manager
	time.Sleep(1 * time.Second) 
	close(rm.UpdateChan)
	for msg := range rm.UpdateChan {
		fmt.Println("[UI UPDATE]:", msg)
	}

	fmt.Println("=========================================")
	fmt.Println("   💤 SERVER GOING TO SLEEP 💤   ")
	fmt.Println("=========================================")
}
