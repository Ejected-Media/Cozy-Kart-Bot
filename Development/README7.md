Here is the code for the Starfield Screensaver.
This is written in pure HTML5 Canvas (JavaScript). It recreates that classic "Warp Speed" look where stars rush toward you. It is lightweight, efficient, and perfect for the "Offline Mode" of Cozy OS.
1. The Code (web/templates/components/starfield.html)
You can save this as a separate file and include it, or paste it directly into your index.html.
<canvas id="starfield" style="position: fixed; top: 0; left: 0; width: 100%; height: 100%; z-index: -1; background: black;"></canvas>

<script>
    const canvas = document.getElementById('starfield');
    const ctx = canvas.getContext('2d');

    let width, height;
    
    // Configuration
    const STAR_COUNT = 800;
    const SPEED = 0.05; // How fast we fly
    
    // The Star Data Structure
    let stars = [];

    function init() {
        resize();
        window.addEventListener('resize', resize);
        
        // Create random stars
        for (let i = 0; i < STAR_COUNT; i++) {
            stars.push({
                x: Math.random() * width - width / 2, // Random X centered
                y: Math.random() * height - height / 2, // Random Y centered
                z: Math.random() * width // Random depth (distance from camera)
            });
        }
        
        requestAnimationFrame(animate);
    }

    function resize() {
        width = window.innerWidth;
        height = window.innerHeight;
        canvas.width = width;
        canvas.height = height;
    }

    function animate() {
        // Clear screen with slight fade for trails (optional, using solid black for classic look)
        ctx.fillStyle = "black";
        ctx.fillRect(0, 0, width, height);
        
        ctx.fillStyle = "white";
        
        // Move the origin to the center of the screen
        const cx = width / 2;
        const cy = height / 2;

        for (let star of stars) {
            // Move star closer to viewer (decrease Z)
            star.z -= width * SPEED;

            // Reset star if it passes the camera (Z <= 0)
            if (star.z <= 0) {
                star.z = width;
                star.x = Math.random() * width - width / 2;
                star.y = Math.random() * height - height / 2;
            }

            // Project 3D coordinates to 2D screen space
            // The logic: (x / z) creates perspective
            const x2d = (star.x / star.z) * width + cx;
            const y2d = (star.y / star.z) * height + cy;

            // Calculate size based on proximity (closer = bigger)
            const size = (1 - star.z / width) * 3;

            // Draw the star
            if (x2d >= 0 && x2d <= width && y2d >= 0 && y2d <= height) {
                ctx.beginPath();
                ctx.arc(x2d, y2d, size, 0, Math.PI * 2);
                ctx.fill();
            }
        }

        requestAnimationFrame(animate);
    }

    // Start the engine!
    init();
</script>

2. How it fits into the UI Layers
To make this feel like an OS and not just a mess of divs, you need to understand the Z-Index Layering.
 * Layer 0 (Bottom): The Starfield Canvas (The "Wallpaper").
 * Layer 1 (Middle): The Twitch Embed (The "App Window").
 * Layer 2 (Top): The HUD Widgets (Clock, Weather, Ticker).
3. Integrating it into your Main Page
In your index.html, you simply wrap your Twitch Player in a div that we can hide/show.
// Inside your updateOSMode() function from before:

if (status.is_racing) {
    // RACE MODE:
    // 1. Hide the canvas (or just let it run behind the video)
    // 2. Show the Twitch Player container
    document.getElementById('twitch-container').style.display = 'block';
} else {
    // OFFLINE MODE:
    // 1. Hide the Twitch Player
    document.getElementById('twitch-container').style.display = 'none';
    // 2. Now the Starfield (which is fixed to the background) is fully visible!
}

4. Customizing the "Vibe"
Since you are in Phoenix and mentioned "After Dark," here are two easy tweaks you can make to the code above:
 * "Desert Night" Mode: Change ctx.fillStyle = "white" to a mix of #ffd700 (Gold) and #ff4500 (Orange) to make the stars look like distant city lights or embers.
 * "Warp Speed" Control: Bind the SPEED variable to your mouse position. When you move your mouse to the right, the stars fly faster!
You now have a dashboard that is alive even when no one is racing. It's a perfect digital lava lamp.
