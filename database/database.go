package database

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
)

type Database struct {
	db *sql.DB
	mu sync.Mutex
}

func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS pages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT NOT NULL UNIQUE,
		content TEXT NOT NULL
	)`)
	if err != nil {
		return nil, fmt.Errorf("error creating pages table: %v", err)
	}

	return &Database{db: db}, nil
}

func (d *Database) InsertPage(url, content string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.db.Exec("INSERT INTO pages (url, content) VALUES (?, ?)", url, content)
	if err != nil {
		return fmt.Errorf("error inserting page into database: %v", err)
	}

	return nil
}

func (d *Database) GetPage(url string) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	var content string
	err := d.db.QueryRow("SELECT content FROM pages WHERE url = ?", url).Scan(&content)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("error getting page from database: %v", err)
	}

	return content, nil
}

func (d *Database) Close() {
	err := d.db.Close()
	if err != nil {
		log.Printf("error closing database: %v", err)
	}
}
