package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Request struct for /api/evaluate endpoint
type Request struct {
	Solution      string `json:"solution"`
	Turns         []Turn `json:"turns"`
	ProposedGuess string `json:"proposed_guess"`
}

type Turn struct {
	Guess    string `json:"guess"`
	Feedback string `json:"feedback"`
}

// Response struct for /api/evaluate endpoint
type Response struct {
	GameStatus         string `json:"game_status"`
	TurnValid          bool   `json:"turn_valid"`
	ShortlistReduction struct {
		Before int     `json:"before"`
		After  int     `json:"after"`
		Ratio  float64 `json:"ratio"`
	} `json:"shortlist_reduction"`
	Feedback string `json:"feedback"`
}

func evaluateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var req Request
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Basic validation
	if len(req.Solution) != 5 || req.Solution != strings.ToLower(req.Solution) {
		http.Error(w, "Invalid solution format", http.StatusBadRequest)
		return
	}

	if req.ProposedGuess != "" && (len(req.ProposedGuess) != 5 || req.ProposedGuess != strings.ToLower(req.ProposedGuess)) {
		http.Error(w, "Invalid guess format", http.StatusBadRequest)
		return
	}

	// Validate all turns
	for _, turn := range req.Turns {
		if len(turn.Guess) != 5 || turn.Guess != strings.ToLower(turn.Guess) {
			http.Error(w, "Invalid guess format in turns", http.StatusBadRequest)
			return
		}
		if len(turn.Feedback) != 5 {
			http.Error(w, "Invalid feedback format", http.StatusBadRequest)
			return
		}
	}

	// Dummy response
	dummyResp := Response{
		GameStatus: "ongoing",
		TurnValid:  true,
		ShortlistReduction: struct {
			Before int     `json:"before"`
			After  int     `json:"after"`
			Ratio  float64 `json:"ratio"`
		}{
			Before: 2309,
			After:  100,
			Ratio:  0.9567,
		},
		Feedback: "G--YY",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dummyResp)
}

func main() {
	http.HandleFunc("/api/evaluate", evaluateHandler)
	http.ListenAndServe(":9111", nil)
}
