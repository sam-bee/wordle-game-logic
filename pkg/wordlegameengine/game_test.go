package wordlegameengine

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if err := LoadWordlists("../../data"); err != nil {
		fmt.Printf("Failed to load wordlists: %v\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func TestNewGame(t *testing.T) {
	solution := mustNewSolution("crane")
	game := NewGame(solution)

	if game.Solution != solution {
		t.Errorf("NewGame().Solution = %v, want %v", game.Solution, solution)
	}
	if len(game.Guesses) != 0 {
		t.Errorf("NewGame().Guesses length = %d, want 0", len(game.Guesses))
	}
	if len(game.Feedbacks) != 0 {
		t.Errorf("NewGame().Feedbacks length = %d, want 0", len(game.Feedbacks))
	}
	if cap(game.Guesses) != MaxGuesses {
		t.Errorf("NewGame().Guesses capacity = %d, want %d", cap(game.Guesses), MaxGuesses)
	}
	if cap(game.Feedbacks) != MaxGuesses {
		t.Errorf("NewGame().Feedbacks capacity = %d, want %d", cap(game.Feedbacks), MaxGuesses)
	}
}

func TestGame_PlayGuess(t *testing.T) {
	solution := mustNewSolution("crane")
	game := NewGame(solution)

	guess := mustNewWord("slate")
	game.PlayGuess(guess)

	if len(game.Guesses) != 1 {
		t.Errorf("after PlayGuess, Guesses length = %d, want 1", len(game.Guesses))
	}
	if game.Guesses[0] != guess {
		t.Errorf("after PlayGuess, Guesses[0] = %v, want %v", game.Guesses[0], guess)
	}
	if len(game.Feedbacks) != 1 {
		t.Errorf("after PlayGuess, Feedbacks length = %d, want 1", len(game.Feedbacks))
	}

	// slate vs crane: s=grey, l=grey, a=green, t=grey, e=green
	expectedFeedback := Feedback{Grey, Grey, Green, Grey, Green}
	if game.Feedbacks[0] != expectedFeedback {
		t.Errorf("after PlayGuess, Feedbacks[0] = %v, want %v", game.Feedbacks[0], expectedFeedback)
	}
}

func TestGame_PlayGuess_Multiple(t *testing.T) {
	solution := mustNewSolution("crane")
	game := NewGame(solution)

	guesses := []string{"slate", "trace", "crane"}
	for _, g := range guesses {
		game.PlayGuess(mustNewWord(g))
	}

	if len(game.Guesses) != 3 {
		t.Errorf("after 3 PlayGuess calls, Guesses length = %d, want 3", len(game.Guesses))
	}
	if len(game.Feedbacks) != 3 {
		t.Errorf("after 3 PlayGuess calls, Feedbacks length = %d, want 3", len(game.Feedbacks))
	}
}

func TestGame_LastFeedback(t *testing.T) {
	t.Run("no guesses", func(t *testing.T) {
		solution := mustNewSolution("crane")
		game := NewGame(solution)

		if game.LastFeedback() != nil {
			t.Errorf("LastFeedback() = %v, want nil", game.LastFeedback())
		}
	})

	t.Run("one guess", func(t *testing.T) {
		solution := mustNewSolution("crane")
		game := NewGame(solution)
		game.PlayGuess(mustNewWord("slate"))

		feedback := game.LastFeedback()
		if feedback == nil {
			t.Fatal("LastFeedback() = nil, want non-nil")
		}

		expectedFeedback := Feedback{Grey, Grey, Green, Grey, Green}
		if *feedback != expectedFeedback {
			t.Errorf("LastFeedback() = %v, want %v", *feedback, expectedFeedback)
		}
	})

	t.Run("multiple guesses returns last", func(t *testing.T) {
		solution := mustNewSolution("crane")
		game := NewGame(solution)
		game.PlayGuess(mustNewWord("slate"))
		game.PlayGuess(mustNewWord("crane"))

		feedback := game.LastFeedback()
		if feedback == nil {
			t.Fatal("LastFeedback() = nil, want non-nil")
		}

		expectedFeedback := Feedback{Green, Green, Green, Green, Green}
		if *feedback != expectedFeedback {
			t.Errorf("LastFeedback() = %v, want %v", *feedback, expectedFeedback)
		}
	})
}

func TestGame_Won(t *testing.T) {
	t.Run("no guesses", func(t *testing.T) {
		solution := mustNewSolution("crane")
		game := NewGame(solution)

		if game.Won() {
			t.Error("Won() = true, want false for no guesses")
		}
	})

	t.Run("incorrect guess", func(t *testing.T) {
		solution := mustNewSolution("crane")
		game := NewGame(solution)
		game.PlayGuess(mustNewWord("slate"))

		if game.Won() {
			t.Error("Won() = true, want false for incorrect guess")
		}
	})

	t.Run("correct guess", func(t *testing.T) {
		solution := mustNewSolution("crane")
		game := NewGame(solution)
		game.PlayGuess(mustNewWord("crane"))

		if !game.Won() {
			t.Error("Won() = false, want true for correct guess")
		}
	})

	t.Run("correct guess after incorrect guesses", func(t *testing.T) {
		solution := mustNewSolution("crane")
		game := NewGame(solution)
		game.PlayGuess(mustNewWord("slate"))
		game.PlayGuess(mustNewWord("trace"))
		game.PlayGuess(mustNewWord("crane"))

		if !game.Won() {
			t.Error("Won() = false, want true for correct guess")
		}
	})

	t.Run("incorrect guess after previous guesses", func(t *testing.T) {
		solution := mustNewSolution("crane")
		game := NewGame(solution)
		game.PlayGuess(mustNewWord("slate"))
		game.PlayGuess(mustNewWord("trace"))

		if game.Won() {
			t.Error("Won() = true, want false for incorrect last guess")
		}
	})
}

func TestGame_SolutionShortlist_Smoke(t *testing.T) {
	// Set up a game with solution "spare" and a pre-filtered shortlist
	game := &Game{
		Solution:  mustNewSolution("spare"),
		Guesses:   []Word{mustNewWord("scare")},
		Feedbacks: []Feedback{{Green, Grey, Green, Green, Green}},
		SolutionShortlist: []Word{
			mustNewWord("scare"),
			mustNewWord("share"),
			mustNewWord("snare"),
			mustNewWord("spare"),
			mustNewWord("stare"),
		},
	}

	// Play "chant" - should get grey-grey-green-grey-grey against "spare"
	game.PlayGuess(mustNewWord("chant"))

	expectedFeedback := Feedback{Grey, Grey, Green, Grey, Grey}
	if game.Feedbacks[1] != expectedFeedback {
		t.Errorf("feedback for 'chant' = %v, want %v", game.Feedbacks[1], expectedFeedback)
	}

	// Only "spare" should remain in shortlist - it's the only word where
	// "chant" produces grey-grey-green-grey-grey
	if len(game.SolutionShortlist) != 1 {
		t.Errorf("SolutionShortlist length = %d, want 1", len(game.SolutionShortlist))
	}
	if len(game.SolutionShortlist) > 0 && game.SolutionShortlist[0] != mustNewWord("spare") {
		t.Errorf("SolutionShortlist[0] = %v, want 'spare'", game.SolutionShortlist[0])
	}
}

func TestGame_ShortlistLength(t *testing.T) {
	solution := mustNewSolution("crane")
	game := NewGame(solution)

	// Initial shortlist should have all allowed solutions
	initialLength := game.ShortlistLength()
	if initialLength == 0 {
		t.Error("initial ShortlistLength() = 0, want > 0")
	}

	// After playing a guess, shortlist should decrease (or stay same)
	game.PlayGuess(mustNewWord("slate"))
	afterLength := game.ShortlistLength()
	if afterLength > initialLength {
		t.Errorf("ShortlistLength() after guess = %d, should be <= %d", afterLength, initialLength)
	}

	// Test with a guess that eliminates many words
	game2 := NewGame(solution)
	initialLength2 := game2.ShortlistLength()
	game2.PlayGuess(mustNewWord("aaaaa"))
	afterLength2 := game2.ShortlistLength()
	if afterLength2 >= initialLength2 {
		t.Errorf("ShortlistLength() after invalid guess = %d, should be < %d", afterLength2, initialLength2)
	}
}

func TestGame_ReplayTurn(t *testing.T) {
	solution := mustNewSolution("crane")
	game := NewGame(solution)

	// Replay a turn with known guess and feedback
	guess := mustNewWord("slate")
	feedback := Feedback{Grey, Grey, Green, Grey, Green}
	game.ReplayTurn(guess, feedback)

	// Verify guess and feedback were recorded
	if len(game.Guesses) != 1 {
		t.Errorf("after ReplayTurn, Guesses length = %d, want 1", len(game.Guesses))
	}
	if game.Guesses[0] != guess {
		t.Errorf("after ReplayTurn, Guesses[0] = %v, want %v", game.Guesses[0], guess)
	}
	if len(game.Feedbacks) != 1 {
		t.Errorf("after ReplayTurn, Feedbacks length = %d, want 1", len(game.Feedbacks))
	}
	if game.Feedbacks[0] != feedback {
		t.Errorf("after ReplayTurn, Feedbacks[0] = %v, want %v", game.Feedbacks[0], feedback)
	}

	// Verify shortlist was updated
	// For "slate" with feedback "--G-G" against "crane", only words with 'a' at pos 2 and 'e' at pos 4 should remain
	shortlistLength := game.ShortlistLength()
	if shortlistLength == len(AllowedSolutions) {
		t.Error("Shortlist should have been updated after ReplayTurn")
	}
}

func TestGame_ReplayTurn_Multiple(t *testing.T) {
	solution := mustNewSolution("crane")
	game := NewGame(solution)

	// Replay multiple turns
	turns := []struct {
		guess    string
		feedback Feedback
	}{
		{"slate", Feedback{Grey, Grey, Green, Grey, Green}},
		{"crane", Feedback{Green, Green, Green, Green, Green}},
	}

	for _, turn := range turns {
		game.ReplayTurn(mustNewWord(turn.guess), turn.feedback)
	}

	if len(game.Guesses) != 2 {
		t.Errorf("after 2 ReplayTurn calls, Guesses length = %d, want 2", len(game.Guesses))
	}
	if len(game.Feedbacks) != 2 {
		t.Errorf("after 2 ReplayTurn calls, Feedbacks length = %d, want 2", len(game.Feedbacks))
	}
}

func mustNewSolution(s string) Solution {
	sol, err := NewSolution(s)
	if err != nil {
		panic(err)
	}
	return sol
}
