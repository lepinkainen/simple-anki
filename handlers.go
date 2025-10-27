package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func respondJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, message string, status int) {
	respondJSON(w, map[string]string{"error": message}, status)
}

// CardsHandler handles /api/cards
func CardsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Get all cards or filter by deck
		deckName := r.URL.Query().Get("deck")
		cards, err := GetAllCards(deckName)
		if err != nil {
			respondError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respondJSON(w, cards, http.StatusOK)

	case "POST":
		// Create new card
		var card Card
		if err := json.NewDecoder(r.Body).Decode(&card); err != nil {
			respondError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if card.Front == "" || card.Back == "" {
			respondError(w, "Front and back are required", http.StatusBadRequest)
			return
		}

		if card.DeckName == "" {
			card.DeckName = "Default"
		}

		if err := CreateCard(&card); err != nil {
			respondError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		respondJSON(w, card, http.StatusCreated)

	default:
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// CardHandler handles /api/cards/{id}
func CardHandler(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/cards/")
	id, err := strconv.Atoi(path)
	if err != nil {
		respondError(w, "Invalid card ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		card, err := GetCard(id)
		if err != nil {
			respondError(w, "Card not found", http.StatusNotFound)
			return
		}
		respondJSON(w, card, http.StatusOK)

	case "PUT":
		var card Card
		if err := json.NewDecoder(r.Body).Decode(&card); err != nil {
			respondError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		card.ID = id
		if err := UpdateCard(&card); err != nil {
			respondError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		respondJSON(w, card, http.StatusOK)

	case "DELETE":
		if err := DeleteCard(id); err != nil {
			respondError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respondJSON(w, map[string]string{"message": "Card deleted"}, http.StatusOK)

	default:
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// DecksHandler handles /api/decks
func DecksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	decks, err := GetDecks()
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, decks, http.StatusOK)
}

// ReviewHandler handles /api/review
func ReviewHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Get due cards for review
		deckName := r.URL.Query().Get("deck")
		limitStr := r.URL.Query().Get("limit")
		limit := 20
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil {
				limit = l
			}
		}

		cards, err := GetDueCards(deckName, limit)
		if err != nil {
			respondError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respondJSON(w, cards, http.StatusOK)

	case "POST":
		// Submit review result
		var result ReviewResult
		if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
			respondError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if result.Score < 1 || result.Score > 4 {
			respondError(w, "Score must be between 1 and 4", http.StatusBadRequest)
			return
		}

		card, err := GetCard(result.CardID)
		if err != nil {
			respondError(w, "Card not found", http.StatusNotFound)
			return
		}

		CalculateNextReview(card, result.Score)

		if err := UpdateCard(card); err != nil {
			respondError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		respondJSON(w, card, http.StatusOK)

	default:
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ImportRequest represents the JSON structure for importing cards
type ImportRequest struct {
	DeckName string `json:"deck_name"`
	Cards    []struct {
		Front string `json:"front"`
		Back  string `json:"back"`
	} `json:"cards"`
}

// ImportHandler handles /api/import
func ImportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var importReq ImportRequest
	if err := json.NewDecoder(r.Body).Decode(&importReq); err != nil {
		respondError(w, "Invalid JSON format: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate deck_name
	if importReq.DeckName == "" {
		respondError(w, "deck_name is required and cannot be empty", http.StatusBadRequest)
		return
	}

	// Validate cards array
	if len(importReq.Cards) == 0 {
		respondError(w, "cards array is required and must contain at least one card", http.StatusBadRequest)
		return
	}

	// Validate and import each card
	importedCount := 0
	for i, cardData := range importReq.Cards {
		// Validate front and back
		if cardData.Front == "" {
			respondError(w, "Card at index "+strconv.Itoa(i)+" has empty 'front' field", http.StatusBadRequest)
			return
		}
		if cardData.Back == "" {
			respondError(w, "Card at index "+strconv.Itoa(i)+" has empty 'back' field", http.StatusBadRequest)
			return
		}

		// Create card
		card := Card{
			DeckName: importReq.DeckName,
			Front:    cardData.Front,
			Back:     cardData.Back,
		}

		if err := CreateCard(&card); err != nil {
			respondError(w, "Failed to import card at index "+strconv.Itoa(i)+": "+err.Error(), http.StatusInternalServerError)
			return
		}

		importedCount++
	}

	// Success response
	respondJSON(w, map[string]interface{}{
		"success":        true,
		"imported_count": importedCount,
		"deck_name":      importReq.DeckName,
		"message":        "Successfully imported " + strconv.Itoa(importedCount) + " cards into deck '" + importReq.DeckName + "'",
	}, http.StatusCreated)
}
