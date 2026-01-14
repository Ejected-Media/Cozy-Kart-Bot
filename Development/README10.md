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
