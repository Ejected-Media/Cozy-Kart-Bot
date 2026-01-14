Since you want both, let's complete the Data Layer (Firestore) first to finish off the backend, and then we'll jump straight into the Android Client (Kotlin) to see how the player actually receives that data.
Part 1: The Data Layer (Go + Firestore)
We need to save the race results so we can generate that "Journal/Scrapbook" visualization later. Since you mentioned Fuzzy-Go (your other project) uses Firestore, this pattern should feel familiar, but with a specific focus on atomic transactions for the currency.
Here is storage.go. It handles saving the RacerSession we created earlier.
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

Part 2: The Android Client (Kotlin)
Now, let's switch to the phone.
Remember the Discord Bot sent a "Magic Link" (cozykart.gg/dashboard?token=xyz)? The Android app needs to intercept that link so it can launch the game and log the user in automatically.
Here is how you handle that Deep Link in Kotlin.
1. AndroidManifest.xml
First, we tell Android: "If anyone clicks a link to cozykart.gg, open THIS app, not Chrome."
<activity android:name=".MainActivity">
    <intent-filter android:autoVerify="true">
        <action android:name="android.intent.action.VIEW" />
        <category android:name="android.intent.category.DEFAULT" />
        <category android:name="android.intent.category.BROWSABLE" />

        <data android:scheme="https" android:host="cozykart.gg" android:pathPrefix="/dashboard" />
    </intent-filter>
</activity>

2. MainActivity.kt
Now, the Kotlin code to grab that token and authenticate.
package com.example.cozykart

import android.content.Intent
import android.net.Uri
import android.os.Bundle
import android.util.Log
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity
import androidx.lifecycle.lifecycleScope
import kotlinx.coroutines.launch

class MainActivity : AppCompatActivity() {

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        // Check if the app was started by the Magic Link
        handleIntent(intent)
    }

    override fun onNewIntent(intent: Intent) {
        super.onNewIntent(intent)
        // Handle cases where the app is already running in the background
        handleIntent(intent)
    }

    private fun handleIntent(intent: Intent) {
        val appLinkAction = intent.action
        val appLinkData: Uri? = intent.data

        if (Intent.ACTION_VIEW == appLinkAction && appLinkData != null) {
            // Extract the JWT token from the URL query parameter
            val token = appLinkData.getQueryParameter("token")

            if (token != null) {
                Log.d("CozyKart", "Magic Token Received: $token")
                authenticateUser(token)
            } else {
                Toast.makeText(this, "Invalid Login Link", Toast.LENGTH_SHORT).show()
            }
        }
    }

    private fun authenticateUser(token: String) {
        // Here we would use Retrofit to send the token to our Go Backend
        // to verify it and establish a persistent WebSocket connection.
        
        lifecycleScope.launch {
            try {
                // Mocking the repository call
                // val userProfile = Repository.login(token)
                
                Toast.makeText(this@MainActivity, "Welcome back, Racer!", Toast.LENGTH_LONG).show()
                
                // Navigate to the "Green Room" Fragment
                // findNavController().navigate(R.id.action_global_greenRoomFragment)
                
            } catch (e: Exception) {
                Log.e("CozyKart", "Login Failed", e)
            }
        }
    }
}

The Full Cycle
 * Discord: User types /race. Bot replies with https://cozykart.gg/dashboard?token=ABC.
 * Android: User taps link. AndroidManifest catches it.
 * Kotlin: MainActivity extracts ABC.
 * Network: Kotlin sends ABC to Go Backend.
 * Go: Verifies ABC using the JWTSecret.
 * Firestore: Go fetches the user's profile and lifetime stats.
 * UI: The "Racer Control Panel" (the first image we drew) loads up with their stats.
This covers the entire loop from the initial visual concept to the database implementation! Is there a specific part of this stack (The UI, The Bot, The Backend, or The App) you want to dive deeper into?
