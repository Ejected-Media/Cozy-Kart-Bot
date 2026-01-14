package com.example.cozykart.ui.greenroom

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

// The possible states of our Green Room screen
sealed class RoomState {
    object Idle : RoomState()                     // "Waiting for players..."
    object Staging : RoomState()                  // "Ready Up!" (Button Active)
    data class Locked(val code: String) : RoomState() // "Code: 8841" (Ignition)
}

class GreenRoomViewModel : ViewModel() {

    // The UI observes this flow to know what to draw
    private val _uiState = MutableStateFlow<RoomState>(RoomState.Idle)
    val uiState: StateFlow<RoomState> = _uiState

    private val _isUserReady = MutableStateFlow(false)
    val isUserReady: StateFlow<Boolean> = _isUserReady

    // Simulate connecting to the backend
    init {
        simulateBackendEvents()
    }

    // Called when user taps the Big Button
    fun onReadyClicked() {
        _isUserReady.value = true
        // In real app: webSocket.send("CONFIRM_READY")
    }

    // This simulates the "RaceManager" sending signals to the phone
    private fun simulateBackendEvents() {
        viewModelScope.launch {
            delay(3000) // Wait 3 seconds...
            _uiState.value = RoomState.Staging // BACKEND SAYS: "STAGING START!"
            
            delay(5000) // Wait 5 seconds...
            // If user readied up, show the code
            if (_isUserReady.value) {
                _uiState.value = RoomState.Locked("8841-COZY")
            }
        }
    }
}
