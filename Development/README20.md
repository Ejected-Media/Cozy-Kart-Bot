No, we haven't written that one yet! We built the Green Room (waiting area), but we still need the Race Screen itself.
This is the "Compact Mode" we visualized earlier. Since the user is likely busy driving (playing the game on a screen), this phone screen acts as a Digital Dashboard or Telemetry Unit. It needs to be high-contrast, easy to read at a glance, and focused entirely on the Money and the Vibes.
Here is the source code for RaceScreen.kt and its accompanying state logic.
1. The Logic (RaceViewModel.kt)
This ViewModel handles the real-time telemetry. It listens for "Pot Updates" and "Vibe Checks" from the backend.
package com.example.cozykart.ui.race

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch
import kotlin.random.Random

data class RaceTelemetry(
    val raceTime: String = "00:00",
    val currentPot: Double = 0.00,
    val splitPercent: Int = 60, // 60% standard
    val isVibeBonusActive: Boolean = false, // The "On Fire" state
    val lap: Int = 1,
    val totalLaps: Int = 3
)

class RaceViewModel : ViewModel() {

    private val _telemetry = MutableStateFlow(RaceTelemetry())
    val telemetry: StateFlow<RaceTelemetry> = _telemetry

    init {
        // In a real app, this would start the WebSocket subscription.
        // Here, we simulate the race data stream.
        simulateRaceData()
    }

    private fun simulateRaceData() {
        viewModelScope.launch {
            val startTime = System.currentTimeMillis()
            var pot = 5.00 // Start with $5 seeded

            while (true) {
                val elapsed = System.currentTimeMillis() - startTime
                val seconds = (elapsed / 1000) % 60
                val minutes = (elapsed / 1000) / 60
                
                // Simulate random donations coming in
                if (Random.nextBoolean()) {
                    pot += 0.50
                }

                // Simulate "Vibe Bonus" triggering after 10 seconds
                val bonusActive = elapsed > 10_000
                val currentSplit = if (bonusActive) 70 else 60

                _telemetry.value = _telemetry.value.copy(
                    raceTime = String.format("%02d:%02d", minutes, seconds),
                    currentPot = pot,
                    isVibeBonusActive = bonusActive,
                    splitPercent = currentSplit
                )

                delay(1000) // Update every second
            }
        }
    }
}

2. The Visuals (RaceScreen.kt)
This uses Jetpack Compose to create a "Heads Up Display" (HUD).
 * Vibe Mode: If isVibeBonusActive is true, the whole interface glows (changes color).
 * Money: The font size is massive because that's what the racer cares about.
<!-- end list -->
package com.example.cozykart.ui.race

import androidx.compose.animation.animateColorAsState
import androidx.compose.animation.core.tween
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import java.text.NumberFormat
import java.util.Locale

// Design System Colors
val HUD_Dark = Color(0xFF121212)
val HUD_Standard = Color(0xFF00E5FF) // Cyan
val HUD_Bonus = Color(0xFFFFD700)    // Gold (Vibe Bonus)

@Composable
fun RaceScreen(viewModel: RaceViewModel) {
    val telemetry by viewModel.telemetry.collectAsState()

    // Animate color change when Vibe Bonus hits
    val accentColor by animateColorAsState(
        targetValue = if (telemetry.isVibeBonusActive) HUD_Bonus else HUD_Standard,
        animationSpec = tween(durationMillis = 1000), 
        label = "VibeColorAnimation"
    )

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(HUD_Dark)
            .padding(16.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.SpaceBetween
    ) {
        
        // --- TOP BAR: TIMER & LAP ---
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween
        ) {
            HUDStat("LAP", "${telemetry.lap}/${telemetry.totalLaps}", accentColor)
            HUDStat("TIME", telemetry.raceTime, accentColor)
        }

        // --- CENTER: THE POT (MAIN FOCUS) ---
        Box(
            modifier = Modifier
                .weight(1f) // Fill available space
                .fillMaxWidth(),
            contentAlignment = Alignment.Center
        ) {
            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                Text(
                    text = "CURRENT PAYOUT",
                    color = Color.Gray,
                    fontSize = 16.sp,
                    letterSpacing = 4.sp
                )
                
                // The Money Counter
                val formattedMoney = NumberFormat.getCurrencyInstance(Locale.US).format(telemetry.currentPot)
                Text(
                    text = formattedMoney,
                    color = accentColor, // Glows Gold when bonus is active
                    fontSize = 80.sp, // Massive font
                    fontWeight = FontWeight.Bold,
                    letterSpacing = (-2).sp
                )
                
                // The Split Percentage Pill
                Spacer(modifier = Modifier.height(16.dp))
                SplitBadge(telemetry.splitPercent, accentColor)
            }
        }

        // --- BOTTOM: VIBE INDICATOR ---
        // This acts as a visual "Engine Light" for the stream status
        Row(
            verticalAlignment = Alignment.CenterVertically,
            modifier = Modifier
                .fillMaxWidth()
                .background(Color.White.copy(alpha = 0.05f), RoundedCornerShape(12.dp))
                .padding(16.dp)
        ) {
            Text(
                text = if (telemetry.isVibeBonusActive) "🔥 VIBE BONUS ACTIVE" else "☁️ STANDARD MODE",
                color = accentColor,
                fontSize = 18.sp,
                fontWeight = FontWeight.Bold,
                modifier = Modifier.weight(1f)
            )
            
            // "Rec" dot to show connection is live
            Box(
                modifier = Modifier
                    .size(12.dp)
                    .background(Color.Red, RoundedCornerShape(50))
            )
        }
    }
}

// Helper Composable for top stats
@Composable
fun HUDStat(label: String, value: String, color: Color) {
    Column(horizontalAlignment = Alignment.Start) {
        Text(text = label, color = Color.Gray, fontSize = 12.sp, fontWeight = FontWeight.Bold)
        Text(text = value, color = color, fontSize = 28.sp, fontWeight = FontWeight.Medium)
    }
}

// Helper Composable for the percentage badge
@Composable
fun SplitBadge(percent: Int, color: Color) {
    Text(
        text = "$percent% SPLIT",
        color = HUD_Dark,
        fontSize = 14.sp,
        fontWeight = FontWeight.Black,
        modifier = Modifier
            .background(color, RoundedCornerShape(50))
            .padding(horizontal = 16.dp, vertical = 8.dp)
    )
}

Why this code works for "Cozy-Kart":
 * Glanceability: The logic prioritizes the currentPot. The font size 80.sp means the racer can look down at their phone for 0.5 seconds and immediately know how much money they are making.
 * Visual Feedback: The animateColorAsState transitions the entire UI from "Cyan" (Calm/Standard) to "Gold" (Hype/Bonus) automatically when the backend detects high chat activity.
 * OLED Saving: The background is HUD_Dark (#121212), which saves battery on AMOLED phone screens—critical if the phone is being used as a second screen for a 2-hour session.
Would you like to see the Sequence Diagram of how the phone switches from GreenRoomScreen to RaceScreen when the race starts?
