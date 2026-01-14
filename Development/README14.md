Based on the "Paycheck" module we designed for the Racer Control Panel (the ticker showing "Current Race Pot" and "Session Total"), here is the backend logic for payout.go.
This Go code handles the calculation of earnings, thread-safe updates for the live ticker, and the specific "Split %" logic mentioned in the UI breakdown.
package payout

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Money is represented in cents to avoid floating point errors.
// Example: 1000 = $10.00
type Money int64

// SplitRule defines how the pot is divided based on league standing or "Vibes".
type SplitRule float64

const (
	StandardSplit SplitRule = 0.60 // Base revenue share
	BonusSplit    SplitRule = 0.70 // "High Vibe" bonus share
)

// RacerSession represents the financial state of a racer currently logged into the Dashboard.
type RacerSession struct {
	RacerID      string
	SessionTotal Money     // Total earned since login (The "Session Total" UI element)
	CurrentPot   Money     // The active pot on the table (The "Current Race Pot" UI element)
	ActiveSplit  SplitRule // The percentage shown in the corner
	mu           sync.RWMutex
}

// NewRacerSession initializes the session when the racer opens the dashboard.
func NewRacerSession(id string) *RacerSession {
	return &RacerSession{
		RacerID:      id,
		SessionTotal: 0,
		CurrentPot:   0,
		ActiveSplit:  StandardSplit,
	}
}

// UpdateLivePot receives real-time events (donations/subs) and updates the "Current Race Pot" ticker.
func (rs *RacerSession) UpdateLivePot(amount Money) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	
	// Add to the pending pot
	rs.CurrentPot += amount
}

// CalculateProjectedEarnings returns what the racer *would* get if the race ended right now.
// Used to update the UI ticker in real-time.
func (rs *RacerSession) CalculateProjectedEarnings() Money {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	// Logic: (Pot * Split)
	val := float64(rs.CurrentPot) * float64(rs.ActiveSplit)
	return Money(val)
}

// FinalizeRace locks in the earnings when the race concludes.
// This moves money from the "Pot" to the "Session Total".
func (rs *RacerSession) FinalizeRace(finished bool) (Money, error) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.CurrentPot == 0 {
		return 0, nil
	}

	if !finished {
		// If they DNF (Did Not Finish), the pot might roll over or be forfeited.
		// For Cozy-Kart, let's say they keep 10% as a consolation.
		consolation := Money(float64(rs.CurrentPot) * 0.10)
		rs.SessionTotal += consolation
		rs.CurrentPot = 0
		return consolation, errors.New("race DNF: consolation payout applied")
	}

	// Calculate final payout
	payout := Money(float64(rs.CurrentPot) * float64(rs.ActiveSplit))
	
	// Update the Session Total (The "Bank" ticker)
	rs.SessionTotal += payout
	
	// Reset the Pot for the next race
	rs.CurrentPot = 0

	return payout, nil
}

// ToggleVibeBonus changes the split percentage.
// This might happen if the "Chat Vibes" meter gets high enough.
func (rs *RacerSession) ToggleVibeBonus(active bool) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if active {
		rs.ActiveSplit = BonusSplit
	} else {
		rs.ActiveSplit = StandardSplit
	}
}

// String format for the UI Ticker
func (m Money) String() string {
	return fmt.Sprintf("$%.2f", float64(m)/100.0)
}

// Mock usage for testing the "Paycheck" UI module
func MockTicker() {
	// 1. Racer logs in
	session := NewRacerSession("racer_01")
	fmt.Printf("Display: Session Total: %s\n", session.SessionTotal)

	// 2. Donations come in during the "Green Room" phase
	session.UpdateLivePot(500) // $5.00
	session.UpdateLivePot(250) // $2.50
	
	// 3. Update the UI Ticker
	fmt.Printf("Ticker: Current Pot: %s | Projected: %s\n", session.CurrentPot, session.CalculateProjectedEarnings())

	// 4. Chat goes wild -> Bonus Split activated
	session.ToggleVibeBonus(true)
	fmt.Printf("Ticker Update (Bonus!): Projected: %s\n", session.CalculateProjectedEarnings())

	// 5. Race Ends
	payout, _ := session.FinalizeRace(true)
	fmt.Printf("Race Over. Payout: %s. New Session Total: %s\n", payout, session.SessionTotal)
}

Key Technical Notes for the "Cozy" Architecture:
 * Integer Math: I used int64 for Money. Never use floats for currency in the actual backend, or you'll end up with missing pennies like in Office Space.
 * Concurrency (sync.RWMutex): Since the dashboard ticker updates in real-time (readers) while donations come in simultaneously (writers), the mutex prevents race conditions.
 * Vibe Mechanic: The ToggleVibeBonus function allows the "Split %" on the dashboard to change dynamically based on chat sentiment.
Would you like me to write the middleware that injects the lobby code (The "Key Ignition" module)?
