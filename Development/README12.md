This dashboard, often called the "Racer Control Panel" (RCP), is the command center that stays open on the racer's computer monitor while they play the game on their phone. It bridges the gap between the game, the stream, and the bank.
Here is a breakdown of the specific functional zones usually found in this interface:
1. The "Green Room" (Queue & Status Module)
 * Location: Usually Top-Center or Top-Left (High visibility).
 * Function: This tells the racer exactly where they stand in the league lineup.
 * Key Data Points:
   * Status Indicator: "On Deck," "Racing," "Cooldown," or "Offline."
   * Countdown Timer: "Race Start in 02:45."
   * The 'Ready Up' Button: A massive, physical-looking button they must click to confirm they are at their desk. If they miss this, they forfeit their spot.
2. The "Key Ignition" (Lobby Code Injection)
 * Location: Dead Center (Pop-up style).
 * Function: This is the secure channel for game setup.
 * How it works: When the race is about to start, the system retrieves the private "Room Code" from the Host's Mario Kart lobby and flashes it here. The racer types this into their phone to join the specific match.
 * Security: This code is hidden from the public stream view to prevent "stream snipers" (unauthorized players) from joining the lobby.
3. The "Paycheck" (Real-Time Earnings Ticker)
 * Location: Top Right (The incentive corner).
 * Function: Gamifies the earning process. instead of checking a bank account, they see a live ticker.
 * Key Data Points:
   * "Current Race Pot": How much money is effectively "on the table" for the current lap based on active viewer donations/subscriptions.
   * "Session Total": How much they have earned since logging in today.
   * "Split %": A subtle indicator showing their current revenue share (e.g., "Standard: 60%" vs "Bonus: 70%").
4. The "Pit Crew" (Stream Health Telemetry)
 * Location: Bottom Left (Technical footer).
 * Function: Ensures the broadcast is valid for league regulations.
 * Key Data Points:
   * Bitrate: Must stay above ~4500kbps to ensure the referee can see the gameplay clearly.
   * Frame Drops: Warns if the stream is laggy.
   * OBS Connection: A simple Green/Red light showing if their video feed is reaching the central server.
5. The "Ref's Ear" (Private Comms)
 * Location: Bottom Right or Slide-out panel.
 * Function: A text-based direct line to the League Administrator.
 * Usage: Used for disputes ("My game crashed," "Lag," etc.) without clogging up the public Twitch chat.
Visualizing the Wireframe
To make this crystal clear for your development team, here is a schematic wireframe showing exactly where these modules sit on the screen.
The Interaction Flow
 * Idle: Dashboard shows "Queue Position: #4".
 * Alert: Dashboard flashes Gold. "You are Up! Click Ready."
 * Action: Racer clicks [READY].
 * Info: Center box reveals: "Room Code: 8841."
 * Race: Dashboard minimizes to "Compact Mode" (showing only Earnings & Stream Health) so it doesn't distract the racer.
 * 
