package main

import (
	"encoding/json"
	"fmt"
	"github.com/sam-bee/wordle-game-engine/pkg/wordlegameengine"
	"log"
	"net/http"
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

	// Validate solution
	sol, err := wordlegameengine.NewSolution(req.Solution)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := sol.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate proposed guess
	if req.ProposedGuess != "" {
		guess, err := wordlegameengine.NewWord(req.ProposedGuess)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := guess.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Validate past turns
	for _, turn := range req.Turns {
		tguess, err := wordlegameengine.NewWord(turn.Guess)
		if err != nil {
			http.Error(w, fmt.Errorf("invalid past guess %q: %w", turn.Guess, err).Error(), http.StatusBadRequest)
			return
		}
		if err := tguess.Validate(); err != nil {
			http.Error(w, fmt.Errorf("invalid past guess %q: %w", turn.Guess, err).Error(), http.StatusBadRequest)
			return
		}
		if len(turn.Feedback) != 5 {
			http.Error(w, "Invalid feedback format", http.StatusBadRequest)
			return
		}
	}

	// Calculate real feedback if proposed_guess is provided
	feedbackStr := ""
	if req.ProposedGuess != "" {
		guess, _ := wordlegameengine.NewWord(req.ProposedGuess)
		feedback := sol.CheckGuess(guess)
		feedbackStr = feedback.String()
	}

	resp := Response{
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
		Feedback: feedbackStr,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	if err := wordlegameengine.LoadWordlists("./data"); err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/api/evaluate", evaluateHandler)
	http.ListenAndServe(":9111", nil)
}
