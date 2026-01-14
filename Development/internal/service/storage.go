package storage

import (
	"context"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"cozy-kart/payout"
)

// RaceDocument is the struct we actually save to the DB.
// We use struct tags to tell Firestore exactly what the fields should be named.
type RaceDocument struct {
	RaceID    string    `firestore:"race_id"`
	Timestamp time.Time `firestore:"timestamp"`
	WinnerID  string    `firestore:"winner_id"`
	TotalPot  int64     `firestore:"total_pot"` // Stored in cents
	Results   []Result  `firestore:"results"`
}

type Result struct {
	RacerID string `firestore:"racer_id"`
	Payout  int64  `firestore:"payout"`
	Rank    int    `firestore:"rank"`
}

type Repository struct {
	Client *firestore.Client
}

func NewRepository(ctx context.Context, projectID string) (*Repository, error) {
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return &Repository{Client: client}, nil
}

// RecordRaceOutcome saves the race data and updates the user's lifetime balance atomically.
func (r *Repository) RecordRaceOutcome(ctx context.Context, raceID string, sessions []*payout.RacerSession) error {
	// 1. Prepare the Race Document
	doc := RaceDocument{
		RaceID:    raceID,
		Timestamp: time.Now(),
		Results:   make([]Result, 0),
	}

	var totalPot int64 = 0

	// 2. Start a Firestore Transaction
	// We do this to ensure we don't save the race result without also updating the user's wallet.
	err := r.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		
		for i, session := range sessions {
			// Convert our internal "Money" type back to int64 for storage
			payoutAmt := int64(session.SessionTotal)
			totalPot += payoutAmt

			// Add to the match history doc
			doc.Results = append(doc.Results, Result{
				RacerID: session.RacerID,
				Payout:  payoutAmt,
				Rank:    i + 1, // Assuming the slice is sorted by rank
			})

			// 3. Update the User's Wallet (The "Ledger")
			// We reference the user's document: /users/{userID}
			userRef := r.Client.Collection("users").Doc(session.RacerID)
			
			// Increment their lifetime earnings
			// Firestore "Increment" is safer than reading old value + adding new value
			tx.Update(userRef, []firestore.Update{
				{Path: "lifetime_earnings", Value: firestore.Increment(payoutAmt)},
				{Path: "races_run", Value: firestore.Increment(1)},
			})
		}

		doc.TotalPot = totalPot

		// 4. Save the Match History
		_, err := tx.Set(r.Client.Collection("races").NewDoc(), doc)
		return err
	})

	if err != nil {
		log.Printf("Failed to record race: %v", err)
		return err
	}

	log.Printf("Race %s recorded successfully in Firestore.", raceID)
	return nil
}
