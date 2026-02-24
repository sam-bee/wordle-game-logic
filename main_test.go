package main

import (
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
