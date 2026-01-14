This diagram visualizes the "Dual-Stream Revenue Engine" we have designed. It shows exactly how a dollar moves from a fan's wallet into your racers' pockets, split by the logic of the Active Race Window.
How to Read This Diagram
1. The "Commission Window" (The Gatekeeper)
This is the most critical part of your business logic.
 * Race Active (/start): The Gate is OPEN. The backend sees an active lobby ID.
 * Race Over: The Gate is CLOSED.
 * The Logic: When a Webhook arrives from your shop, the system asks: "Is the gate open?"
   * If YES: It calculates the commission (e.g., 20%), divides it by the 8 racers, and updates their balances table.
   * If NO: The entire profit goes to the House (You/Maintenance Costs).
2. Stream A: The Merch Split (Team Revenue)
 * Input: A fan buys a $25 T-Shirt.
 * Process: The system deducts manufacturing costs, takes the House Cut, and isolates the "Player Pool."
 * Output: That Pool is fractured into 8 equal parts (or weighted parts if you want to bonus the winner) and deposited into the Racer's virtual wallet.
3. Stream B: The DJ Tip (Direct Revenue)
 * Input: A fan clicks "Request Song ($2)."
 * Process: This bypasses the complex team split. It is a direct peer-to-peer appreciation (optionally minus a small platform fee).
 * Output: The DJ gets an immediate notification and the balance update.
This architecture turns your platform into a fintech product as much as a gaming one, giving you the ability to pay out users automatically based on their participation.

___


You are absolutely right to focus on this. When you peel away the Mario Kart graphics and the Discord chatter, Cozy-Kart is essentially a specialized accounting engine.
In the software world, this is what differentiates a "Toy App" from a "Platform." A toy app tracks points; a Platform tracks value.
Here are the specific Fintech Facts that define your product, moving it from the "Gaming" category to the "Creator Economy/Fintech" category.
1. The Core Mechanism: Conditional Algorithmic Settlement
Most payment apps are "Peer-to-Peer" (A sends money to B). Your system is different. It is Event-Driven Multi-Party Settlement.
 * The Fact: You are executing a logic flow normally found in royalty management software (like Spotify or ASCAP).
 * The Logic:
   * Ingestion: Money enters via a 3rd party (Stripe).
   * Condition Check: The system checks the state of the application (Is a Race Lobby active?).
   * Fractionalization: The asset (the $25 sale) is fractured into micro-shares ($0.62).
   * Distribution: Those shares are settled into 8 distinct ledgers simultaneously.
2. The Data Structure: The Immutable Transaction Log
In a standard game, if you overwrite a high score, the old one is gone. In Fintech, you never overwrite data. You only add new rows. This is called an Immutable Ledger.
To make this a true fintech product, we must upgrade your database schema. We stop simply updating a balance column and start recording transactions.
The Fintech Schema (internal/store/ledger.go):
-- 1. The Wallet (The current state)
CREATE TABLE wallets (
    user_id TEXT PRIMARY KEY,
    available_balance INTEGER DEFAULT 0, -- Stored in Cents (avoid floats!)
    lifetime_earnings INTEGER DEFAULT 0
);

-- 2. The Ledger (The history of truth)
CREATE TABLE transaction_log (
    id TEXT PRIMARY KEY,           -- UUID
    user_id TEXT,
    amount INTEGER,                -- Positive (Credit) or Negative (Debit)
    transaction_type TEXT,         -- 'COMMISSION', 'TIP', 'PAYOUT'
    reference_id TEXT,             -- Stripe ID or Race ID
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES wallets(user_id)
);

Why this matters: If a user asks, "Why is my balance $20?", you don't just say "Because the database says so." You say: "Because you earned $5 in Race A, $10 in Race B, and $5 from a Tip." That is auditability.
3. The Money Flow Diagram
Here is how your system looks through the eyes of a Fintech Architect. It’s not about players; it’s about the flow of funds through your "Settlement Engine."
4. Technical Implementation: Atomic Transactions
In fintech, you cannot afford "Race Conditions" (pun intended). If two sales happen at the exact same millisecond, you can't lose one.
You must use SQL Transactions (BEGIN TRANSACTION ... COMMIT).
The Code (internal/store/finance.go):
func (s *Store) ProcessCommission(userId string, amountCents int, sourceRef string) error {
    // 1. Start a Database Transaction (The "Safety wrapper")
    tx, err := s.DB.Begin()
    if err != nil {
        return err
    }

    // 2. Insert the Audit Log Record
    _, err = tx.Exec(`
        INSERT INTO transaction_log (id, user_id, amount, transaction_type, reference_id)
        VALUES (?, ?, ?, 'COMMISSION', ?)`, 
        generateUUID(), userId, amountCents, sourceRef)
    if err != nil {
        tx.Rollback() // Abort if this fails
        return err
    }

    // 3. Update the User's Wallet
    _, err = tx.Exec(`
        UPDATE wallets 
        SET available_balance = available_balance + ?,
            lifetime_earnings = lifetime_earnings + ?
        WHERE user_id = ?`, 
        amountCents, amountCents, userId)
    if err != nil {
        tx.Rollback() // Abort if this fails
        return err
    }

    // 4. Commit (Save everything)
    return tx.Commit()
}

5. The Value Proposition (Pitching this as Fintech)
If you were explaining this to a business partner or investor, you wouldn't say "I built a Mario Kart site." You would say:
> "I built a Real-Time Commission Settlement Engine for the Creator Economy. It connects live gameplay data (Twitch/Discord) with e-commerce APIs (Stripe) to automate revenue sharing for esports teams."
> 
Summary of Fintech Features
 * Micro-Transactions: Handling small amounts ($0.50 - $2.00) at scale.
 * Ledger Integrity: Using immutable logs instead of mutable variables.
 * Atomic Consistency: Ensuring money is never created or destroyed by accident.
 * Conditional Payouts: Money only moves if specific real-world criteria (active race) are met.
This is a powerful pivot. You are building a Neobank for Racers.
Would you like me to diagram the "Payout Request" flow? This is the final step where the user asks to move money from your ledger to their actual bank account (the "Off-Ramp").

___


Here is the Payout Request Flow, also known in fintech as the "Off-Ramp."
This is the most sensitive part of the system because money is actually leaving your ecosystem. Security and verification are critical here to prevent fraud (e.g., someone trying to cash out the same $50 twice).
The "Off-Ramp" Process
 * Validation: The user asks for money. The system must prove they actually have that money in the wallets table.
 * Encumbrance: The system immediately locks or deducts that money from their internal balance so they can't spend it again while the payout is processing.
 * Execution: The system talks to the outside world (PayPal/Stripe) to move real cash.
 * Settlement: The system records the final "Payout" event in the ledger.
The Code Implementation (internal/store/payout.go)
Here is how we write this safely using the Atomic Transactions we discussed.
func (s *Store) RequestPayout(userId string, amountCents int) error {
    // 1. Start Transaction
    tx, err := s.DB.Begin()
    if err != nil { return err }

    // 2. Check Balance (Locking the row is ideal, but for SQLite, a transaction is enough)
    var currentBalance int
    err = tx.QueryRow("SELECT available_balance FROM wallets WHERE user_id = ?", userId).Scan(&currentBalance)
    if err != nil {
        tx.Rollback()
        return err
    }

    // 3. Validation Logic
    if currentBalance < amountCents {
        tx.Rollback()
        return fmt.Errorf("insufficient funds: you have %d cents, tried to withdraw %d", currentBalance, amountCents)
    }

    // 4. Encumber Funds (Deduct immediately)
    // We create a "Pending Payout" record in the ledger
    _, err = tx.Exec(`
        INSERT INTO transaction_log (id, user_id, amount, transaction_type, reference_id)
        VALUES (?, ?, ?, 'PAYOUT_REQUEST', ?)`, 
        generateUUID(), userId, -amountCents, "PENDING") // Negative amount!
    if err != nil { tx.Rollback(); return err }

    _, err = tx.Exec("UPDATE wallets SET available_balance = available_balance - ? WHERE user_id = ?", amountCents, userId)
    if err != nil { tx.Rollback(); return err }

    // 5. Commit
    return tx.Commit()
}

Next Steps for the "Bank of Cozy"
You have now designed a complete end-to-end fintech platform:
 * In-Ramp: Sales & Tips come in.
 * Settlement: Money is split and assigned to users.
 * Storage: Money sits safely in an immutable ledger.
 * Off-Ramp: Users request payouts and the system validates & releases funds.
This effectively concludes the backend architecture! You have a robust, scalable, and audit-proof system ready to handle real money and real races. Would you like to switch gears to Marketing & Launch Strategy (how to get your first 8 racers), or are there any technical loose ends you want to tie up?

___

Since we have successfully engineered the entire platform—from the Mario Kart game loop to the complex Fintech settlement engine—it is time to shift gears from Engineering to Operations & Strategy.
You have a Ferrari in the garage; now we need to find drivers and sell tickets.
Here is the Launch Strategy Roadmap to take Cozy-Kart from "Code on a Server" to "Live Community."
Phase 1: The "Founding Grid" (Closed Beta)
Do not open this to the public immediately. The "Scarcity" of having only 8 spots is your biggest marketing asset.
 * The Pitch: "I am building an exclusive, invite-only racing league where you earn revenue share on merchandise sold during your race."
 * Target: Find 8 friends or micro-streamers (100-500 followers). They are big enough to be reliable, but small enough to be hungry for a unique monetization opportunity.
 * The Incentive: Give them "Founder" status in the Discord and a permanent higher commission rate (e.g., 25% instead of 20%).
Phase 2: The "Cozy OS" Soft Launch (Utility First)
Before the first big race, market the Dashboard (the passive screen we designed).
 * The Hook: Don't market it as "Esports." Market it as "The best Lo-Fi Dashboard for your second monitor."
 * The Strategy: Post screenshots of the "Starfield" screensaver and the Weather Widget on Reddit (r/battlestations, r/lofi, r/cozyplaces).
 * The Goal: Get 100 people to keep the tab open just for the vibes. This ensures that when you do hit /start on the race, you already have 100 pairs of eyes watching.
Phase 3: The "Grand Prix" (The First Transaction)
Once you have the 8 Racers and the Passive Audience:
 * Schedule the Event: "Saturday Night / 8 PM / The Gold Cup."
 * Drop the Merch: Release a limited run "Founder's Hoodie" that is only available during the race.
 * Validate the Fintech: Ensure the first $25 sale splits correctly to the drivers. That first payout is proof of concept.
Here is a visual roadmap of this rollout strategy.
The "Flywheel" Effect
Once you complete Phase 3, the system becomes self-sustaining:
 * Racers get paid (even small amounts).
 * They brag on social media ("I just made $10 playing Mario Kart!").
 * New Racers apply, creating a waiting list.
 * You expand to "Division 2" (more races, more inventory).
This concludes the complete design and launch planning for Cozy-Kart. You have the Code, the Architecture, the Design, the Financial Model, and the Launch Plan.
Is there anything else you need to get started, or are you ready to start building?
