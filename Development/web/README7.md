Here is the Dockerfile to package the entire "Cozy-Kart" stack.
We will use a Multi-Stage Build. This is a best practice that keeps your final image tiny.
 * The Builder Stage: A heavy container with all the Go compilers and tools.
 * The Final Stage: A tiny "Alpine" Linux container (approx. 15MB) that contains only your compiled binary and the HTML templates.
1. The Dockerfile
Create this file in the root of your project (cozy-kart/Dockerfile).
# ----------------------------
# STAGE 1: The Builder
# ----------------------------
FROM golang:1.23-alpine AS builder

# Install git (sometimes needed for go mod download)
RUN apk add --no-cache git

WORKDIR /app

# 1. Download Dependencies first (Caching layer)
# If you change code but not dependencies, Docker skips this step (Fast builds!)
COPY go.mod go.sum ./
RUN go mod download

# 2. Copy the source code
COPY . .

# 3. Build the application
# -o main: Output file name
# cmd/cozy-kart/main.go: Entry point
RUN go build -o cozy-kart cmd/cozy-kart/main.go


# ----------------------------
# STAGE 2: The Runner (Production)
# ----------------------------
FROM alpine:latest

# Install CA Certificates so we can talk to Firestore (HTTPS)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 1. Copy the Binary from the Builder stage
COPY --from=builder /app/cozy-kart .

# 2. IMPORTANT: Copy the Web Templates & Static Files
# The Go binary reads these files from disk at runtime.
COPY --from=builder /app/web ./web

# 3. Expose the port
EXPOSE 8080

# 4. Run it
CMD ["./cozy-kart"]

2. The .dockerignore file
Create a .dockerignore file to prevent copying unnecessary files (like your local git history or secret keys) into the image.
.git
.env
*.md
bin/
tmp/

3. How to Deploy (Deployment Cheat Sheet)
Since you mentioned Render, Railway, or a Raspberry Pi, here is how this Dockerfile works for them:
Option A: Raspberry Pi (Self-Hosted)
Since the Pi uses an ARM processor, Docker handles this automatically.
# 1. Build the image
docker build -t cozy-kart .

# 2. Run it (Passing your keys as Environment Variables)
docker run -d \
  -p 8080:8080 \
  -e DISCORD_TOKEN="your_token_here" \
  -e DISCORD_APP_ID="your_id" \
  -e GCP_PROJECT_ID="your_gcp_project" \
  cozy-kart

Option B: Render / Railway (Cloud)
 * Push your code to GitHub.
 * Connect your repo to Render/Railway.
 * They will detect the Dockerfile automatically.
 * Go to the "Environment Variables" section of their dashboard and paste your .env contents there.
 * Done. They handle the build and hosting.
4. Project Wrap-Up
We have successfully designed the full architecture for Cozy-Kart:
 * Mobile (Android):
   * Green Room: Chatty, status-based UI.
   * Race HUD: High-contrast, OLED-friendly, massive fonts.
 * Discord Bot:
   * Traffic Cop: Routes commands (/host open, /host start).
   * Gateway: Directs players to the mobile app via Magic Links.
 * Backend (Go):
   * Logic: RaceManager holds the state in memory for speed.
   * Comms: WebSockets push updates 2x/second.
   * Storage: Repository pattern saving to Firestore.
 * Web Overlay:
   * HTMX + Alpine: Zero-build frontend for the stream overlay.
   * Templates: Server-side rendering for "The Pot" and "Leaderboard."
Would you like me to synthesize this entire conversation into a single Markdown "README.md" file that you can save as the documentation for the repo?
