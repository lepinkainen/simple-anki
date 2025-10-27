package main

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type Card struct {
	ID         int       `json:"id"`
	DeckName   string    `json:"deck_name"`
	Front      string    `json:"front"`
	Back       string    `json:"back"`
	Ease       float64   `json:"ease"`
	Interval   int       `json:"interval"`
	NextReview time.Time `json:"next_review"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ReviewResult struct {
	CardID int    `json:"card_id"`
	Score  int    `json:"score"` // 1=Again, 2=Hard, 3=Good, 4=Easy
}

func InitDB(dbPath string) error {
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS cards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		deck_name TEXT NOT NULL,
		front TEXT NOT NULL,
		back TEXT NOT NULL,
		ease REAL DEFAULT 2.5,
		interval INTEGER DEFAULT 0,
		next_review DATETIME DEFAULT CURRENT_TIMESTAMP,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_deck_name ON cards(deck_name);
	CREATE INDEX IF NOT EXISTS idx_next_review ON cards(next_review);
	`

	_, err = db.Exec(schema)
	return err
}

func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

func CreateCard(card *Card) error {
	result, err := db.Exec(
		`INSERT INTO cards (deck_name, front, back, ease, interval, next_review)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		card.DeckName, card.Front, card.Back, 2.5, 0, time.Now(),
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	card.ID = int(id)
	return nil
}

func GetCard(id int) (*Card, error) {
	card := &Card{}
	err := db.QueryRow(
		`SELECT id, deck_name, front, back, ease, interval, next_review, created_at, updated_at
		 FROM cards WHERE id = ?`,
		id,
	).Scan(&card.ID, &card.DeckName, &card.Front, &card.Back, &card.Ease, &card.Interval, &card.NextReview, &card.CreatedAt, &card.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return card, nil
}

func GetAllCards(deckName string) ([]Card, error) {
	var rows *sql.Rows
	var err error

	if deckName == "" {
		rows, err = db.Query(
			`SELECT id, deck_name, front, back, ease, interval, next_review, created_at, updated_at
			 FROM cards ORDER BY created_at DESC`,
		)
	} else {
		rows, err = db.Query(
			`SELECT id, deck_name, front, back, ease, interval, next_review, created_at, updated_at
			 FROM cards WHERE deck_name = ? ORDER BY created_at DESC`,
			deckName,
		)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []Card
	for rows.Next() {
		var card Card
		err := rows.Scan(&card.ID, &card.DeckName, &card.Front, &card.Back, &card.Ease, &card.Interval, &card.NextReview, &card.CreatedAt, &card.UpdatedAt)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	return cards, nil
}

func GetDueCards(deckName string, limit int) ([]Card, error) {
	var rows *sql.Rows
	var err error

	if deckName == "" {
		rows, err = db.Query(
			`SELECT id, deck_name, front, back, ease, interval, next_review, created_at, updated_at
			 FROM cards WHERE next_review <= ? ORDER BY next_review LIMIT ?`,
			time.Now(), limit,
		)
	} else {
		rows, err = db.Query(
			`SELECT id, deck_name, front, back, ease, interval, next_review, created_at, updated_at
			 FROM cards WHERE deck_name = ? AND next_review <= ? ORDER BY next_review LIMIT ?`,
			deckName, time.Now(), limit,
		)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []Card
	for rows.Next() {
		var card Card
		err := rows.Scan(&card.ID, &card.DeckName, &card.Front, &card.Back, &card.Ease, &card.Interval, &card.NextReview, &card.CreatedAt, &card.UpdatedAt)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	return cards, nil
}

func GetDecks() ([]string, error) {
	rows, err := db.Query(`SELECT DISTINCT deck_name FROM cards ORDER BY deck_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decks []string
	for rows.Next() {
		var deck string
		if err := rows.Scan(&deck); err != nil {
			return nil, err
		}
		decks = append(decks, deck)
	}

	return decks, nil
}

func UpdateCard(card *Card) error {
	_, err := db.Exec(
		`UPDATE cards SET deck_name = ?, front = ?, back = ?, ease = ?, interval = ?, next_review = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		card.DeckName, card.Front, card.Back, card.Ease, card.Interval, card.NextReview, card.ID,
	)
	return err
}

func DeleteCard(id int) error {
	_, err := db.Exec(`DELETE FROM cards WHERE id = ?`, id)
	return err
}

// Simple SM-2 algorithm implementation
func CalculateNextReview(card *Card, score int) {
	// score: 1=Again, 2=Hard, 3=Good, 4=Easy

	if score < 3 {
		// Failed: reset interval
		card.Interval = 0
		card.Ease = max(1.3, card.Ease-0.2)
		card.NextReview = time.Now().Add(1 * time.Minute) // Review again in 1 minute
	} else {
		// Passed: increase interval
		if card.Interval == 0 {
			card.Interval = 1
		} else if card.Interval == 1 {
			card.Interval = 6
		} else {
			card.Interval = int(float64(card.Interval) * card.Ease)
		}

		// Adjust ease factor
		if score == 3 {
			// Good - no change to ease
		} else if score == 4 {
			// Easy - increase ease
			card.Ease = min(card.Ease+0.15, 2.5)
		} else if score == 2 {
			// Hard - decrease ease
			card.Ease = max(1.3, card.Ease-0.15)
		}

		card.NextReview = time.Now().Add(time.Duration(card.Interval) * 24 * time.Hour)
	}
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
