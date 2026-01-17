package wordlegameengine

type Solution Word

type TileColor int8

const (
	Grey TileColor = iota
	Yellow
	Green
)

type Feedback [WordLength]TileColor

func (s *Solution) CheckGuess(guess Word) Feedback {
	var feedback Feedback
	var used [WordLength]bool

	// First pass: mark greens
	for i := 0; i < WordLength; i++ {
		if guess[i] == s[i] {
			feedback[i] = Green
			used[i] = true
		}
	}

	// Second pass: mark yellows
	for i := 0; i < WordLength; i++ {
		if feedback[i] == Green {
			continue
		}
		for j := 0; j < WordLength; j++ {
			if !used[j] && guess[i] == s[j] {
				feedback[i] = Yellow
				used[j] = true
				break
			}
		}
	}

	return feedback
}
