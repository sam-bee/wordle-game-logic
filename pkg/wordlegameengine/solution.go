package wordlegameengine

import "fmt"

type Solution Word

func errNotInSolutions(s string) error {
	return fmt.Errorf("%q not in allowed solutions", s)
}

func NewSolution(s string) (Solution, error) {
	var sol Solution
	if err := parseWord(s, sol[:]); err != nil {
		return Solution{}, err
	}
	return sol, nil
}

func (s Solution) String() string {
	return string(s[:])
}

func (s *Solution) Validate() error {
	str := s.String()
	if err := validateCharacters(str); err != nil {
		return err
	}
	if !isInWordlist(str, AllowedSolutions) {
		return errNotInSolutions(str)
	}
	return nil
}

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

func (f Feedback) String() string {
	// Convert each TileColor to its character representation:
	// Green -> 'G', Yellow -> 'Y', Grey -> '-'
	result := make([]byte, WordLength)
	for i, color := range f {
		switch color {
		case Green:
			result[i] = 'G'
		case Yellow:
			result[i] = 'Y'
		default:
			result[i] = '-'
		}
	}
	return string(result)
}
