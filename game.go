package wordlegameengine

import "math/rand/v2"

const MaxGuesses = 6

type Game struct {
	Solution          Solution
	Guesses           []Word
	Feedbacks         []Feedback
	PossibleSolutions []Word
}

func NewGame(solution Solution) *Game {
	return &Game{
		Solution:          solution,
		Guesses:           make([]Word, 0, MaxGuesses),
		Feedbacks:         make([]Feedback, 0, MaxGuesses),
		PossibleSolutions: append([]Word{}, AllowedSolutions...),
	}
}

func NewRandomGame() *Game {
	idx := rand.IntN(len(AllowedSolutions))
	solution := Solution(AllowedSolutions[idx])
	return NewGame(solution)
}

func (g *Game) PlayGuess(guess Word) {
	feedback := g.Solution.CheckGuess(guess)
	g.Guesses = append(g.Guesses, guess)
	g.Feedbacks = append(g.Feedbacks, feedback)
	g.updatePossibleSolutions()
}

func (g *Game) updatePossibleSolutions() {
	var remaining []Word
	for _, candidate := range g.PossibleSolutions {
		if g.matchesFeedback(candidate) {
			remaining = append(remaining, candidate)
		}
	}
	g.PossibleSolutions = remaining
}

func (g *Game) matchesFeedback(candidate Word) bool {
	candidateSolution := Solution(candidate)

	for i, guess := range g.Guesses {
		expectedFeedback := g.Feedbacks[i]
		actualFeedback := candidateSolution.CheckGuess(guess)

		if actualFeedback != expectedFeedback {
			return false
		}
	}

	return true
}

func (g *Game) LastFeedback() *Feedback {
	if len(g.Feedbacks) == 0 {
		return nil
	}
	return &g.Feedbacks[len(g.Feedbacks)-1]
}

func (g *Game) Won() bool {
	feedback := g.LastFeedback()
	if feedback == nil {
		return false
	}
	for i := 0; i < WordLength; i++ {
		if feedback[i] != Green {
			return false
		}
	}
	return true
}
