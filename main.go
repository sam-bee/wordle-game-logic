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

	// Check for first turn cache
	var cacheKey wordlegameengine.CacheKey
	haveFirstTurn := len(req.Turns) > 0
	cached := false
	var cachedShortlist []wordlegameengine.Word

	if haveFirstTurn {
		firstGuess, err := wordlegameengine.NewWord(req.Turns[0].Guess)
		if err != nil {
			http.Error(w, fmt.Errorf("invalid past guess %q: %w", req.Turns[0].Guess, err).Error(), http.StatusBadRequest)
			return
		}
		if err := firstGuess.Validate(); err != nil {
			http.Error(w, fmt.Errorf("invalid past guess %q: %w", req.Turns[0].Guess, err).Error(), http.StatusBadRequest)
			return
		}
		firstFeedback, err := wordlegameengine.ParseFeedback(req.Turns[0].Feedback)
		if err != nil {
			http.Error(w, fmt.Errorf("invalid feedback %q: %w", req.Turns[0].Feedback, err).Error(), http.StatusBadRequest)
			return
		}
		cacheKey = wordlegameengine.MakeCacheKey(firstGuess, firstFeedback)
		cachedShortlist, cached = wordlegameengine.FirstTurnCache.Get(cacheKey)
	}

	// Create game based on cache status
	var game *wordlegameengine.Game

	if haveFirstTurn && cached {
		// Cache hit: Create game with cached shortlist
		game = wordlegameengine.NewGameWithShortlist(sol, cachedShortlist)
	} else {
		// Cache miss or no turns: Create game normally
		game = wordlegameengine.NewGame(sol)
	}

	// Replay turns (skip first turn if cache hit)
	startIdx := 0
	if haveFirstTurn && cached {
		startIdx = 1 // Skip first turn - already in cached shortlist
	}

	for i := startIdx; i < len(req.Turns); i++ {
		turn := req.Turns[i]
		guess, err := wordlegameengine.NewWord(turn.Guess)
		if err != nil {
			http.Error(w, fmt.Errorf("invalid past guess %q: %w", turn.Guess, err).Error(), http.StatusBadRequest)
			return
		}
		if err := guess.Validate(); err != nil {
			http.Error(w, fmt.Errorf("invalid past guess %q: %w", turn.Guess, err).Error(), http.StatusBadRequest)
			return
		}
		feedback, err := wordlegameengine.ParseFeedback(turn.Feedback)
		if err != nil {
			http.Error(w, fmt.Errorf("invalid feedback %q: %w", turn.Feedback, err).Error(), http.StatusBadRequest)
			return
		}
		game.ReplayTurn(guess, feedback)
	}

	// Cache on first-turn miss
	if haveFirstTurn && !cached && len(req.Turns) == 1 {
		shortlistCopy := make([]wordlegameengine.Word, len(game.SolutionShortlist))
		copy(shortlistCopy, game.SolutionShortlist)
		wordlegameengine.FirstTurnCache.Put(cacheKey, shortlistCopy)
	}

	// Get shortlist length BEFORE playing proposed guess
	before := game.ShortlistLength()

	// Calculate real feedback and shortlist reduction if proposed_guess is provided
	feedbackStr := ""
	after := before // If no proposed guess, after = before (no reduction)
	if req.ProposedGuess != "" {
		guess, _ := wordlegameengine.NewWord(req.ProposedGuess)
		feedback := sol.CheckGuess(guess)
		feedbackStr = feedback.String()
		game.PlayGuess(guess)
		after = game.ShortlistLength()
	}

	// Calculate ratio (handle division by zero)
	ratio := 0.0
	if before > 0 {
		ratio = 1.0 - (float64(after) / float64(before))
	}

	resp := Response{
		GameStatus: "ongoing",
		TurnValid:  true,
		ShortlistReduction: struct {
			Before int     `json:"before"`
			After  int     `json:"after"`
			Ratio  float64 `json:"ratio"`
		}{
			Before: before,
			After:  after,
			Ratio:  ratio,
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

	// Initialize the B-tree cache
	wordlegameengine.InitCache()

	http.HandleFunc("/api/evaluate", evaluateHandler)
	http.ListenAndServe(":9111", nil)
}
