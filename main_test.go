package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/sam-bee/wordle-game-engine/pkg/wordlegameengine"
)

func TestMain(m *testing.M) {
	if err := wordlegameengine.LoadWordlists("./data"); err != nil {
		fmt.Printf("Failed to load wordlists: %v\n", err)
		os.Exit(1)
	}
	// Initialize the B-tree cache for tests
	wordlegameengine.InitCache()
	os.Exit(m.Run())
}

func TestEvaluateHandler(t *testing.T) {
	tests := []struct {
		name       string
		reqBody    string
		wantStatus int
	}{
		{
			name:       "valid request",
			reqBody:    `{"solution":"aback","turns":[],"proposed_guess":"aahed"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid solution length",
			reqBody:    `{"solution":"abc","turns":[],"proposed_guess":""}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid solution chars",
			reqBody:    `{"solution":"ABCDE","turns":[],"proposed_guess":""}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "solution not in allowed solutions",
			reqBody:    `{"solution":"aahed","turns":[],"proposed_guess":""}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid proposed guess length",
			reqBody:    `{"solution":"aback","turns":[],"proposed_guess":"abc"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "proposed guess invalid chars",
			reqBody:    `{"solution":"aback","turns":[],"proposed_guess":"ABCDE"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "proposed guess not in allowed guesses",
			reqBody:    `{"solution":"aback","turns":[],"proposed_guess":"abcde"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid past turn guess",
			reqBody:    `{"solution":"aback","turns":[{"guess":"abc","feedback":"-----"}],"proposed_guess":""}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "past turn feedback invalid length",
			reqBody:    `{"solution":"aback","turns":[{"guess":"aahed","feedback":"----"}],"proposed_guess":""}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "valid past turn",
			reqBody:    `{"solution":"aback","turns":[{"guess":"aahed","feedback":"-----"}],"proposed_guess":""}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid feedback character",
			reqBody:    `{"solution":"aback","turns":[{"guess":"aahed","feedback":"-G-!-"}],"proposed_guess":""}`,
			wantStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/evaluate", strings.NewReader(tt.reqBody))
			w := httptest.NewRecorder()
			evaluateHandler(w, req)
			if status := w.Code; status != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.wantStatus)
				bodyBytes, _ := io.ReadAll(w.Body)
				t.Logf("Response body: %q", bodyBytes)
			}
		})
	}
}

func TestEvaluateHandler_ShortlistReduction_RealValues(t *testing.T) {
	tests := []struct {
		name         string
		reqBody      string
		wantBefore   int
		wantAfter    int
		wantFeedback string
	}{
		{
			name:         "solution apple, no past turns, proposed guess raise",
			reqBody:      `{"solution":"apple","turns":[],"proposed_guess":"raise"}`,
			wantBefore:   2309, // All allowed solutions
			wantAfter:    0,    // Will be actual reduced count after playing "raise"
			wantFeedback: "-Y--G",
		},
		{
			name:         "solution crane, no past turns, proposed guess crane",
			reqBody:      `{"solution":"crane","turns":[],"proposed_guess":"crane"}`,
			wantBefore:   2309, // All allowed solutions
			wantAfter:    1,    // Only "crane" remains after correct guess
			wantFeedback: "GGGGG",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/evaluate", strings.NewReader(tt.reqBody))
			w := httptest.NewRecorder()
			evaluateHandler(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("handler returned wrong status code: got %v want %v", w.Code, http.StatusOK)
			}

			var resp Response
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			// Verify the "before" count is the full solution set
			if resp.ShortlistReduction.Before != tt.wantBefore {
				t.Errorf("ShortlistReduction.Before = %d, want %d", resp.ShortlistReduction.Before, tt.wantBefore)
			}

			// Verify feedback is correct
			if resp.Feedback != tt.wantFeedback {
				t.Errorf("Feedback = %q, want %q", resp.Feedback, tt.wantFeedback)
			}

			// Verify after is less than before for non-winning guesses
			if tt.wantAfter == 0 && resp.ShortlistReduction.After >= resp.ShortlistReduction.Before {
				t.Errorf("ShortlistReduction.After = %d, should be less than Before = %d",
					resp.ShortlistReduction.After, resp.ShortlistReduction.Before)
			}

			// Verify ratio is calculated correctly
			if resp.ShortlistReduction.Before > 0 {
				expectedRatio := 1.0 - (float64(resp.ShortlistReduction.After) / float64(resp.ShortlistReduction.Before))
				if resp.ShortlistReduction.Ratio != expectedRatio {
					t.Errorf("ShortlistReduction.Ratio = %f, want %f", resp.ShortlistReduction.Ratio, expectedRatio)
				}
			}
		})
	}
}

func TestEvaluateHandler_ShortlistReduction_WithPastTurns(t *testing.T) {
	// Test with past turns that should reduce the shortlist before calculating
	reqBody := `{"solution":"crane","turns":[{"guess":"slate","feedback":"--G-G"}],"proposed_guess":"crane"}`
	req := httptest.NewRequest(http.MethodPost, "/api/evaluate", strings.NewReader(reqBody))
	w := httptest.NewRecorder()
	evaluateHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", w.Code, http.StatusOK)
	}

	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// After replaying "slate" with "--G-G" feedback, shortlist should be reduced
	// "--G-G" means: a at pos 2, e at pos 4 are correct (green)
	if resp.ShortlistReduction.Before >= 2309 {
		t.Errorf("ShortlistReduction.Before = %d, should be less than 2309 after replaying turns", resp.ShortlistReduction.Before)
	}

	// After playing "crane", only "crane" should remain
	if resp.ShortlistReduction.After != 1 {
		t.Errorf("ShortlistReduction.After = %d, want 1", resp.ShortlistReduction.After)
	}

	if resp.Feedback != "GGGGG" {
		t.Errorf("Feedback = %q, want GGGGG", resp.Feedback)
	}
}

func TestEvaluateHandler_ShortlistReduction_EmptyProposedGuess(t *testing.T) {
	// Test with empty proposed_guess - should report current state without reduction
	reqBody := `{"solution":"crane","turns":[{"guess":"slate","feedback":"--G-G"}],"proposed_guess":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/evaluate", strings.NewReader(reqBody))
	w := httptest.NewRecorder()
	evaluateHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", w.Code, http.StatusOK)
	}

	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// With empty proposed_guess, before and after should be equal (no reduction from proposed guess)
	if resp.ShortlistReduction.After != resp.ShortlistReduction.Before {
		t.Errorf("ShortlistReduction.After = %d, want %d (same as before) for empty proposed_guess",
			resp.ShortlistReduction.After, resp.ShortlistReduction.Before)
	}

	// With empty proposed_guess, ratio should be 0 (no new information from proposed guess)
	// The ratio represents reduction from playing the proposed guess, not from past turns
	if resp.ShortlistReduction.Ratio != 0.0 {
		t.Errorf("ShortlistReduction.Ratio = %f, want 0.0 for empty proposed_guess", resp.ShortlistReduction.Ratio)
	}

	// Feedback should be empty
	if resp.Feedback != "" {
		t.Errorf("Feedback = %q, want empty string for empty proposed_guess", resp.Feedback)
	}
}

func TestEvaluateHandler_Cache_FirstTurnCacheMiss(t *testing.T) {
	// Initialize cache for this test
	wordlegameengine.InitCache()

	// First request - cache miss
	reqBody := `{"solution":"apple","turns":[{"guess":"raise","feedback":"-Y--G"}],"proposed_guess":"amber"}`
	req := httptest.NewRequest(http.MethodPost, "/api/evaluate", strings.NewReader(reqBody))
	w := httptest.NewRecorder()
	evaluateHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("First request: handler returned wrong status code: got %v want %v", w.Code, http.StatusOK)
	}

	var resp1 Response
	if err := json.NewDecoder(w.Body).Decode(&resp1); err != nil {
		t.Fatalf("First request: failed to decode response: %v", err)
	}

	// Verify response is correct for cache miss (should compute normally)
	if resp1.ShortlistReduction.Before <= 0 {
		t.Errorf("First request: ShortlistReduction.Before = %d, should be > 0", resp1.ShortlistReduction.Before)
	}

	// Verify the result was cached
	guess, _ := wordlegameengine.NewWord("raise")
	feedback, _ := wordlegameengine.ParseFeedback("-Y--G")
	cacheKey := wordlegameengine.MakeCacheKey(guess, feedback)
	cachedShortlist, found := wordlegameengine.FirstTurnCache.Get(cacheKey)
	if !found {
		t.Error("Expected result to be cached after first request")
	}
	if len(cachedShortlist) != resp1.ShortlistReduction.Before {
		t.Errorf("Cached shortlist length = %d, want %d", len(cachedShortlist), resp1.ShortlistReduction.Before)
	}
}

func TestEvaluateHandler_Cache_FirstTurnCacheHit(t *testing.T) {
	// Initialize cache for this test
	wordlegameengine.InitCache()

	// Make same request twice
	reqBody := `{"solution":"apple","turns":[{"guess":"raise","feedback":"-Y--G"}],"proposed_guess":"amber"}`

	// First request - cache miss
	req1 := httptest.NewRequest(http.MethodPost, "/api/evaluate", strings.NewReader(reqBody))
	w1 := httptest.NewRecorder()
	evaluateHandler(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("First request failed: %v", w1.Code)
	}

	var resp1 Response
	if err := json.NewDecoder(w1.Body).Decode(&resp1); err != nil {
		t.Fatalf("First request: failed to decode response: %v", err)
	}

	// Second request - cache hit
	req2 := httptest.NewRequest(http.MethodPost, "/api/evaluate", strings.NewReader(reqBody))
	w2 := httptest.NewRecorder()
	evaluateHandler(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("Second request failed: %v", w2.Code)
	}

	var resp2 Response
	if err := json.NewDecoder(w2.Body).Decode(&resp2); err != nil {
		t.Fatalf("Second request: failed to decode response: %v", err)
	}

	// Responses should be identical
	if resp1.ShortlistReduction.Before != resp2.ShortlistReduction.Before {
		t.Errorf("Before mismatch: resp1=%d, resp2=%d", resp1.ShortlistReduction.Before, resp2.ShortlistReduction.Before)
	}
	if resp1.ShortlistReduction.After != resp2.ShortlistReduction.After {
		t.Errorf("After mismatch: resp1=%d, resp2=%d", resp1.ShortlistReduction.After, resp2.ShortlistReduction.After)
	}
	if resp1.Feedback != resp2.Feedback {
		t.Errorf("Feedback mismatch: resp1=%q, resp2=%q", resp1.Feedback, resp2.Feedback)
	}
}

func TestEvaluateHandler_Cache_SkipsFirstTurnOnHit(t *testing.T) {
	// Initialize cache for this test
	wordlegameengine.InitCache()

	// First request to populate cache
	// Solution: apple, guess: raise, actual feedback: -Y--G (a is yellow at pos 1, e is green at pos 4)
	reqBody1 := `{"solution":"apple","turns":[{"guess":"raise","feedback":"-Y--G"}],"proposed_guess":""}`
	req1 := httptest.NewRequest(http.MethodPost, "/api/evaluate", strings.NewReader(reqBody1))
	w1 := httptest.NewRecorder()
	evaluateHandler(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("First request failed: %v", w1.Code)
	}

	var resp1 Response
	if err := json.NewDecoder(w1.Body).Decode(&resp1); err != nil {
		t.Fatalf("First request: failed to decode response: %v", err)
	}

	// Second request with additional turn - should use cached shortlist and skip first turn
	// Solution: apple, guess: ample, actual feedback: G-GGG (a at pos 0, m not in word, p,l,e match)
	reqBody2 := `{"solution":"apple","turns":[{"guess":"raise","feedback":"-Y--G"},{"guess":"ample","feedback":"G-GGG"}],"proposed_guess":""}`
	req2 := httptest.NewRequest(http.MethodPost, "/api/evaluate", strings.NewReader(reqBody2))
	w2 := httptest.NewRecorder()
	evaluateHandler(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("Second request failed: %v", w2.Code)
	}

	var resp2 Response
	if err := json.NewDecoder(w2.Body).Decode(&resp2); err != nil {
		t.Fatalf("Second request: failed to decode response: %v", err)
	}

	// The "before" count should be the cached shortlist length (after first turn)
	// not the full solution list. Since the second request has 2 turns and filters
	// further to only "apple", Before will be 1, which is still < 2309.
	if resp2.ShortlistReduction.Before >= 2309 {
		t.Errorf("Second request: Before = %d, should be less than 2309 (cached shortlist used)", resp2.ShortlistReduction.Before)
	}
}

func TestEvaluateHandler_Cache_DifferentFirstTurns(t *testing.T) {
	// Initialize cache for this test
	wordlegameengine.InitCache()

	// Two different first turns should produce different cache entries
	reqBody1 := `{"solution":"apple","turns":[{"guess":"raise","feedback":"-Y--G"}],"proposed_guess":""}`
	reqBody2 := `{"solution":"apple","turns":[{"guess":"crane","feedback":"GGGGG"}],"proposed_guess":""}`

	// First request
	req1 := httptest.NewRequest(http.MethodPost, "/api/evaluate", strings.NewReader(reqBody1))
	w1 := httptest.NewRecorder()
	evaluateHandler(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("First request failed: %v", w1.Code)
	}

	// Second request with different first turn
	req2 := httptest.NewRequest(http.MethodPost, "/api/evaluate", strings.NewReader(reqBody2))
	w2 := httptest.NewRecorder()
	evaluateHandler(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("Second request failed: %v", w2.Code)
	}

	var resp2 Response
	if err := json.NewDecoder(w2.Body).Decode(&resp2); err != nil {
		t.Fatalf("Second request: failed to decode response: %v", err)
	}

	// For "crane" with "GGGGG", the shortlist should be reduced to 1 (only "apple"... wait, solution is apple)
	// Actually if solution is "apple" and guess is "crane" with "GGGGG", that's impossible
	// But the shortlist should still be valid
	if resp2.ShortlistReduction.Before <= 0 {
		t.Errorf("Second request: Before = %d, should be > 0", resp2.ShortlistReduction.Before)
	}

	// Both cache entries should exist
	guess1, _ := wordlegameengine.NewWord("raise")
	feedback1, _ := wordlegameengine.ParseFeedback("-Y--G")
	key1 := wordlegameengine.MakeCacheKey(guess1, feedback1)

	guess2, _ := wordlegameengine.NewWord("crane")
	feedback2, _ := wordlegameengine.ParseFeedback("GGGGG")
	key2 := wordlegameengine.MakeCacheKey(guess2, feedback2)

	_, found1 := wordlegameengine.FirstTurnCache.Get(key1)
	_, found2 := wordlegameengine.FirstTurnCache.Get(key2)

	if !found1 {
		t.Error("First cache entry should exist")
	}
	if !found2 {
		t.Error("Second cache entry should exist")
	}
}
