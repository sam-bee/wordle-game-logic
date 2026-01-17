package wordlegameengine

import (
	"testing"
)

func TestNewWord(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"valid word", "hello", nil},
		{"valid word all a", "aaaaa", nil},
		{"valid word all z", "zzzzz", nil},
		{"too short", "hell", ErrInvalidLength},
		{"too long", "helloo", ErrInvalidLength},
		{"empty string", "", ErrInvalidLength},
		{"uppercase letter", "Hello", ErrInvalidCharacter},
		{"contains number", "hell0", ErrInvalidCharacter},
		{"contains space", "hell ", ErrInvalidCharacter},
		{"contains hyphen", "he-lo", ErrInvalidCharacter},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			word, err := NewWord(tt.input)
			if err != tt.wantErr {
				t.Errorf("NewWord(%q) error = %v, want %v", tt.input, err, tt.wantErr)
			}
			if err == nil && word.String() != tt.input {
				t.Errorf("NewWord(%q).String() = %q, want %q", tt.input, word.String(), tt.input)
			}
		})
	}
}

func TestWord_String(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"hello", "hello"},
		{"world", "world"},
		{"aaaaa", "aaaaa"},
		{"zzzzz", "zzzzz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			word, _ := NewWord(tt.input)
			if got := word.String(); got != tt.input {
				t.Errorf("Word.String() = %q, want %q", got, tt.input)
			}
		})
	}
}

func TestWord_Validate(t *testing.T) {
	// Set up test wordlist
	oldGuesses := AllowedGuesses
	defer func() { AllowedGuesses = oldGuesses }()

	AllowedGuesses = []Word{}
	for _, s := range []string{"apple", "berry", "crane", "delta", "eager"} {
		w, _ := NewWord(s)
		AllowedGuesses = append(AllowedGuesses, w)
	}

	tests := []struct {
		name    string
		word    Word
		wantErr error
	}{
		{"valid word in list (first)", mustNewWord("apple"), nil},
		{"valid word in list (middle)", mustNewWord("crane"), nil},
		{"valid word in list (last)", mustNewWord("eager"), nil},
		{"valid word not in list", mustNewWord("zebra"), ErrNotInWordlist},
		{"invalid uppercase", Word{'H', 'e', 'l', 'l', 'o'}, ErrInvalidCharacter},
		{"invalid null byte", Word{0, 'e', 'l', 'l', 'o'}, ErrInvalidCharacter},
		{"invalid number", Word{'h', 'e', 'l', 'l', '0'}, ErrInvalidCharacter},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.word.Validate()
			if err != tt.wantErr {
				t.Errorf("Word.Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func mustNewWord(s string) Word {
	w, err := NewWord(s)
	if err != nil {
		panic(err)
	}
	return w
}
