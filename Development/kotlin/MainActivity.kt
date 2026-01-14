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
