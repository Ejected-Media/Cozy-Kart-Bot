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

