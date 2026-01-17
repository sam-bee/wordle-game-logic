package wordlegameengine

import (
	"testing"
)

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

func TestGame_AddGuess(t *testing.T) {
	solution := mustNewSolution("crane")
	game := NewGame(solution)

	guess := mustNewWord("slate")
	game.AddGuess(guess)

	if len(game.Guesses) != 1 {
		t.Errorf("after AddGuess, Guesses length = %d, want 1", len(game.Guesses))
	}
	if game.Guesses[0] != guess {
		t.Errorf("after AddGuess, Guesses[0] = %v, want %v", game.Guesses[0], guess)
	}
	if len(game.Feedbacks) != 1 {
		t.Errorf("after AddGuess, Feedbacks length = %d, want 1", len(game.Feedbacks))
	}

	// slate vs crane: s=grey, l=grey, a=green, t=grey, e=green
	expectedFeedback := Feedback{Grey, Grey, Green, Grey, Green}
	if game.Feedbacks[0] != expectedFeedback {
		t.Errorf("after AddGuess, Feedbacks[0] = %v, want %v", game.Feedbacks[0], expectedFeedback)
	}
}

func TestGame_AddGuess_Multiple(t *testing.T) {
	solution := mustNewSolution("crane")
	game := NewGame(solution)

	guesses := []string{"slate", "trace", "crane"}
	for _, g := range guesses {
		game.AddGuess(mustNewWord(g))
	}

	if len(game.Guesses) != 3 {
		t.Errorf("after 3 AddGuess calls, Guesses length = %d, want 3", len(game.Guesses))
	}
	if len(game.Feedbacks) != 3 {
		t.Errorf("after 3 AddGuess calls, Feedbacks length = %d, want 3", len(game.Feedbacks))
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
		game.AddGuess(mustNewWord("slate"))

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
		game.AddGuess(mustNewWord("slate"))
		game.AddGuess(mustNewWord("crane"))

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
		game.AddGuess(mustNewWord("slate"))

		if game.Won() {
			t.Error("Won() = true, want false for incorrect guess")
		}
	})

	t.Run("correct guess", func(t *testing.T) {
		solution := mustNewSolution("crane")
		game := NewGame(solution)
		game.AddGuess(mustNewWord("crane"))

		if !game.Won() {
			t.Error("Won() = false, want true for correct guess")
		}
	})

	t.Run("correct guess after incorrect guesses", func(t *testing.T) {
		solution := mustNewSolution("crane")
		game := NewGame(solution)
		game.AddGuess(mustNewWord("slate"))
		game.AddGuess(mustNewWord("trace"))
		game.AddGuess(mustNewWord("crane"))

		if !game.Won() {
			t.Error("Won() = false, want true for correct guess")
		}
	})

	t.Run("incorrect guess after previous guesses", func(t *testing.T) {
		solution := mustNewSolution("crane")
		game := NewGame(solution)
		game.AddGuess(mustNewWord("slate"))
		game.AddGuess(mustNewWord("trace"))

		if game.Won() {
			t.Error("Won() = true, want false for incorrect last guess")
		}
	})
}

func mustNewSolution(s string) Solution {
	sol, err := NewSolution(s)
	if err != nil {
		panic(err)
	}
	return sol
}
