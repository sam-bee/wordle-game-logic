package wordlegameengine

import (
	"math/rand/v2"
	"sync"
)

const MaxGuesses = 6
const numWorkers = 16

type Game struct {
	Solution          Solution
	Guesses           []Word
	Feedbacks         []Feedback
	SolutionShortlist []Word
}

func NewGame(solution Solution) *Game {
	return &Game{
		Solution:          solution,
		Guesses:           make([]Word, 0, MaxGuesses),
		Feedbacks:         make([]Feedback, 0, MaxGuesses),
		SolutionShortlist: append([]Word{}, AllowedSolutions...),
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
	g.updateSolutionShortlist()
}

func (g *Game) updateSolutionShortlist() {

	// Update the game's solution shortlist. Do this by looping over the previous shortlist,
	// assessing Words with game.matchesFeedback(), and keeping the remaining possibilities.

	// prevShortlist should be populated, g.SolutionShortlist should be empty for now

	prevShortlist := g.SolutionShortlist
	g.SolutionShortlist = make([]Word, 0)

	if len(prevShortlist) == 0 {
		return
	}

	// Prep channels for communicating with worker pool that checks words to see if they should be on the shortlist

	prevShortlistCh := make(chan Word, len(prevShortlist))
	newShortlistCh := make(chan Word, len(prevShortlist))

	// Spin up worker pool to test candidate words for the solution shortlist

	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(in <-chan Word, out chan<- Word) {
			defer wg.Done()
			for candidate := range in {
				if g.matchesFeedback(candidate) {
					out <- candidate
				}
			}
		}(prevShortlistCh, newShortlistCh)
	}

	// Put the previous shortlist on the channel going into the worker pool. The workers will then assess whether the
	// words belong on the shortlist.

	for _, candidate := range prevShortlist {
		prevShortlistCh <- candidate
	}
	close(prevShortlistCh)

	// Wait till the worker pool is ready using the waitgroup, then close its output channel for writing
	wg.Wait()
	close(newShortlistCh)

	// Read all the words off the worker pool's output channel, and save to g.SolutionShortlist
	for word := range newShortlistCh {
		g.SolutionShortlist = append(g.SolutionShortlist, word)
	}
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

func (g *Game) ShortlistLength() int {
	return len(g.SolutionShortlist)
}

func (g *Game) ReplayTurn(guess Word, feedback Feedback) {
	g.Guesses = append(g.Guesses, guess)
	g.Feedbacks = append(g.Feedbacks, feedback)
	g.updateSolutionShortlist()
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
