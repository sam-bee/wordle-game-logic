package wordlegameengine

const MaxGuesses = 6

type Game struct {
	Solution  Solution
	Guesses   []Word
	Feedbacks []Feedback
}

func NewGame(solution Solution) *Game {
	return &Game{
		Solution:  solution,
		Guesses:   make([]Word, 0, MaxGuesses),
		Feedbacks: make([]Feedback, 0, MaxGuesses),
	}
}

func (g *Game) AddGuess(guess Word) {
	feedback := g.Solution.CheckGuess(guess)
	g.Guesses = append(g.Guesses, guess)
	g.Feedbacks = append(g.Feedbacks, feedback)
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
