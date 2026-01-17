package wordlegameengine

import (
	"testing"
)

func TestSolution_CheckGuess(t *testing.T) {
	tests := []struct {
		name     string
		solution string
		guess    string
		want     Feedback
	}{
		{
			name:     "test where all feedback tiles should be grey",
			solution: "raise",
			guess:    "clout",
			want:     Feedback{Grey, Grey, Grey, Grey, Grey},
		},
		{
			name:     "test where all feedback tiles should be green",
			solution: "raise",
			guess:    "raise",
			want:     Feedback{Green, Green, Green, Green, Green},
		},
		{
			name:     "test use of yellow tiles appropriate for repeat letters",
			solution: "asses",
			guess:    "sassy",
			want:     Feedback{Yellow, Yellow, Green, Yellow, Grey},
		},
		{
			name:     "test appropriate use of grey tiles in feedback absent repeat letters in solution",
			solution: "waves",
			guess:    "sassy",
			want:     Feedback{Yellow, Green, Grey, Grey, Grey},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			solution := Solution(mustNewWord(tt.solution))
			guess := mustNewWord(tt.guess)
			got := solution.CheckGuess(guess)
			if got != tt.want {
				t.Errorf("Solution(%q).CheckGuess(%q) = %v, want %v",
					tt.solution, tt.guess, feedbackString(got), feedbackString(tt.want))
			}
		})
	}
}

func feedbackString(f Feedback) string {
	colors := []rune{'⬜', '🟨', '🟩'}
	result := make([]rune, WordLength)
	for i, c := range f {
		result[i] = colors[c]
	}
	return string(result)
}
