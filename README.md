# Simple Anki - Language Learning Flashcards

A lightweight, web-based flashcard application for language learning with spaced repetition. Built with Go + SQLite backend and vanilla JavaScript frontend (no build step required).

## Features

- **Simple Flashcard System**: Create and study flashcards with front/back content
- **Spaced Repetition (SRS)**: Uses simplified SM-2 algorithm for optimal learning
- **Deck Organization**: Organize cards into different decks (e.g., "Spanish Vocabulary", "French Verbs")
- **Web Interface**: Clean, responsive UI accessible from any browser
- **No Build Step**: Frontend uses vanilla HTML/CSS/JS that's embedded in the Go binary
- **Lightweight**: Single binary with embedded SQLite database

## Quick Start

### Prerequisites

- Go 1.16 or later
- SQLite3

### Installation

1. Clone the repository:
```bash
git clone https://github.com/lepinkainen/simple-anki.git
cd simple-anki
```

2. Install dependencies:
```bash
go get github.com/mattn/go-sqlite3
```

3. Build the application:
```bash
go build -o simple-anki
```

4. Run the server:
```bash
./simple-anki
```

5. Open your browser to `http://localhost:8080`

### Command Line Options

```bash
./simple-anki -port 8080 -db flashcards.db
```

Options:
- `-port`: Server port (default: 8080)
- `-db`: Path to SQLite database file (default: flashcards.db)

## Usage

### Adding Cards

1. Click the "Add Cards" tab
2. Enter a deck name (e.g., "Spanish Vocabulary")
3. Enter the front of the card (e.g., "Hello")
4. Enter the back of the card (e.g., "Hola")
5. Click "Add Card"

### Studying Cards

1. Click the "Study" tab
2. Select a deck or choose "All Decks"
3. Click on a card to flip it and see the answer
4. Rate your knowledge:
   - **Again** (1): Didn't remember - review again soon
   - **Hard** (2): Barely remembered - review sooner
   - **Good** (3): Remembered correctly - standard interval
   - **Easy** (4): Very easy - longer interval

### Managing Cards

1. Click the "Manage" tab
2. View all cards and their next review dates
3. Filter by deck
4. Delete cards as needed

## Data Format

### Database Schema

The application uses SQLite with the following schema:

```sql
CREATE TABLE cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    deck_name TEXT NOT NULL,
    front TEXT NOT NULL,
    back TEXT NOT NULL,
    ease REAL DEFAULT 2.5,           -- Ease factor for SRS (1.3-2.5)
    interval INTEGER DEFAULT 0,      -- Days until next review
    next_review DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Card Object (JSON)

```json
{
  "id": 1,
  "deck_name": "Spanish Vocabulary",
  "front": "Hello",
  "back": "Hola",
  "ease": 2.5,
  "interval": 0,
  "next_review": "2025-10-27T10:00:00Z",
  "created_at": "2025-10-27T10:00:00Z",
  "updated_at": "2025-10-27T10:00:00Z"
}
```

### Field Descriptions

- **id**: Unique identifier for the card
- **deck_name**: Name of the deck (used for organization)
- **front**: Front side of the card (question/word to learn)
- **back**: Back side of the card (answer/translation)
- **ease**: SRS ease factor (affects interval growth, range: 1.3-2.5)
- **interval**: Days until next review (0 = new card)
- **next_review**: Timestamp when card should be reviewed next
- **created_at**: When the card was created
- **updated_at**: When the card was last modified

## REST API

### Endpoints

#### Get All Cards
```
GET /api/cards?deck=DeckName
```
Returns all cards, optionally filtered by deck.

#### Create Card
```
POST /api/cards
Content-Type: application/json

{
  "deck_name": "Spanish",
  "front": "Good morning",
  "back": "Buenos días"
}
```

#### Get Single Card
```
GET /api/cards/{id}
```

#### Update Card
```
PUT /api/cards/{id}
Content-Type: application/json

{
  "deck_name": "Spanish",
  "front": "Good morning",
  "back": "Buenos días"
}
```

#### Delete Card
```
DELETE /api/cards/{id}
```

#### Get All Decks
```
GET /api/decks
```
Returns list of all deck names.

#### Get Due Cards
```
GET /api/review?deck=DeckName&limit=20
```
Returns cards that are due for review.

#### Submit Review
```
POST /api/review
Content-Type: application/json

{
  "card_id": 1,
  "score": 3
}
```
Scores: 1=Again, 2=Hard, 3=Good, 4=Easy

## Spaced Repetition Algorithm

The app uses a simplified SM-2 algorithm:

- **New cards**: Start with 0 interval
- **Score < 3** (Again/Hard): Reset interval to 0, decrease ease
- **Score >= 3** (Good/Easy): Increase interval based on ease factor
  - First review: 1 day
  - Second review: 6 days
  - Subsequent: `interval * ease`
- **Ease adjustments**:
  - Hard (2): -0.15
  - Good (3): no change
  - Easy (4): +0.15
  - Failed: -0.2
  - Minimum: 1.3, Maximum: 2.5

## Future Enhancements (Not Yet Implemented)

- LLM integration for image-to-flashcard conversion
- Bulk import from CSV/JSON
- Statistics and progress tracking
- Card editing in the UI
- Audio pronunciation support
- Image attachments for cards

## License

See LICENSE file for details.