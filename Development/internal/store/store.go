package store

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	DB *sql.DB
}

type Player struct {
	DiscordID   string
	MKTUsername string
	AvatarURL   string
	Status      string
	JoinedAt    time.Time
}

func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Create tables if they don't exist
	schema := `
	CREATE TABLE IF NOT EXISTS players (
		discord_id TEXT PRIMARY KEY,
		mkt_username TEXT,
		avatar_url TEXT
	);
	CREATE TABLE IF NOT EXISTS queue (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		discord_id TEXT,
		status TEXT DEFAULT 'QUEUED', -- QUEUED, INVITED
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}

	return &Store{DB: db}, nil
}

func (s *Store) AddToQueue(discordID, mktName, avatar string) error {
	// 1. Upsert Player (Update name/avatar if they changed)
	_, err := s.DB.Exec(`
		INSERT INTO players (discord_id, mkt_username, avatar_url) VALUES (?, ?, ?)
		ON CONFLICT(discord_id) DO UPDATE SET mkt_username=excluded.mkt_username, avatar_url=excluded.avatar_url`,
		discordID, mktName, avatar)
	if err != nil {
		return err
	}

	// 2. Check if already in queue
	var count int
	s.DB.QueryRow("SELECT COUNT(*) FROM queue WHERE discord_id = ? AND status = 'QUEUED'", discordID).Scan(&count)
	if count > 0 {
		return fmt.Errorf("already_queued")
	}

	// 3. Add to Queue
	_, err = s.DB.Exec("INSERT INTO queue (discord_id, status) VALUES (?, 'QUEUED')", discordID)
	return err
}

// GetActiveGrid fetches up to 8 players for the N64 grid
func (s *Store) GetActiveGrid() ([]Player, error) {
	rows, err := s.DB.Query(`
		SELECT p.discord_id, p.mkt_username, p.avatar_url, q.status, q.created_at
		FROM queue q
		JOIN players p ON q.discord_id = p.discord_id
		WHERE q.status IN ('QUEUED', 'INVITED')
		ORDER BY q.created_at ASC
		LIMIT 8
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grid []Player
	for rows.Next() {
		var p Player
		if err := rows.Scan(&p.DiscordID, &p.MKTUsername, &p.AvatarURL, &p.Status, &p.JoinedAt); err != nil {
			log.Println("Row scan error:", err)
			continue
		}
		grid = append(grid, p)
	}
	return grid, nil
}
