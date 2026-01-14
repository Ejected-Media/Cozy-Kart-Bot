package com.example.cozykart.ui.greenroom

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

// Cozy Color Palette
val CozyOrange = Color(0xFFFFA500)
val DarkSlate = Color(0xFF2D2D2D)
val SuccessGreen = Color(0xFF00C853)

@Composable
fun GreenRoomScreen(viewModel: GreenRoomViewModel) {
    val state by viewModel.uiState.collectAsState()
    val isReady by viewModel.isUserReady.collectAsState()

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(DarkSlate) // Dark mode background
            .padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        
        // 1. STATUS INDICATOR (Top of screen)
        StatusHeader(state)

        Spacer(modifier = Modifier.height(48.dp))

        // 2. THE BIG BUTTON (Center stage)
        // Only show button if we are in Staging phase
        if (state is RoomState.Staging || state is RoomState.Locked) {
            ReadyButton(
                clicked = isReady,
                onClick = { viewModel.onReadyClicked() }
            )
        }

        Spacer(modifier = Modifier.height(32.dp))

        // 3. LOBBY CODE REVEAL (The "Ignition")
        if (state is RoomState.Locked) {
            val code = (state as RoomState.Locked).code
            Text(
                text = "LOBBY CODE: $code",
                color = CozyOrange,
                fontSize = 24.sp,
                fontWeight = FontWeight.Bold,
                modifier = Modifier
                    .background(Color.Black.copy(alpha = 0.3f), RoundedCornerShape(8.dp))
                    .padding(16.dp)
            )
        }
    }
}

@Composable
fun StatusHeader(state: RoomState) {
    val statusText = when (state) {
        is RoomState.Idle -> "Waiting for Host..."
        is RoomState.Staging -> "ARE YOU HERE?"
        is RoomState.Locked -> "START YOUR ENGINES"
    }
    
    Text(
        text = statusText,
        color = Color.White,
        fontSize = 20.sp,
        letterSpacing = 2.sp
    )
}

@Composable
fun ReadyButton(clicked: Boolean, onClick: () -> Unit) {
    val buttonColor = if (clicked) SuccessGreen else CozyOrange
    val buttonText = if (clicked) "LOCKED IN" else "READY UP"

    Button(
        onClick = onClick,
        enabled = !clicked, // Disable after clicking once
        shape = CircleShape, // Make it a big round button
        colors = ButtonDefaults.buttonColors(containerColor = buttonColor),
        modifier = Modifier.size(200.dp) // Massive size
    ) {
        Text(
            text = buttonText,
            fontSize = 24.sp,
            fontWeight = FontWeight.Black
        )
    }
}
