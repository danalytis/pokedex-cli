package main

import "testing"

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    " hello world ",
			expected: []string{"hello", "world"},
		},
		// TODO: add more cases
	}

	for _, c := range cases {
		actual := cleanInput(c.input)

		if len(actual) != len(c.expected) {
			t.Errorf("length mismatch: got %d, want %d for input %q", len(actual), len(c.expected), c.input)
			continue
		}
		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]
			if word != expectedWord {
				t.Errorf("word mismatch at index %d: got %q, want %q for input %q", i, actual[i], c.expected[i], c.input)
			}
		}
	}
}

// TODO: implement
// Additional cleanInput tests
func TestCleanInput_EmptyString(t *testing.T) {
	// Test empty input returns empty slice
}

func TestCleanInput_OnlyWhitespace(t *testing.T) {
	// Test "   " returns empty slice
}

func TestCleanInput_SingleWord(t *testing.T) {
	// Test "hello" returns ["hello"]
}

func TestCleanInput_MultipleSpaces(t *testing.T) {
	// Test "hello    world" returns ["hello", "world"]
}

func TestCleanInput_TabsAndNewlines(t *testing.T) {
	// Test input with \t and \n characters
}

// CLI Command tests (if you have command functions)
func TestCommandHelp(t *testing.T) {
	// Test help command functionality
}

func TestCommandExit(t *testing.T) {
	// Test exit command functionality
}

func TestCommandMap(t *testing.T) {
	// Test map command functionality
}

func TestCommandMapb(t *testing.T) {
	// Test mapb command functionality
}

func TestCommandExplore(t *testing.T) {
	// Test explore command with valid area
}

func TestCommandCatch(t *testing.T) {
	// Test catch command functionality
}

func TestCommandInspect(t *testing.T) {
	// Test inspect command functionality
}

func TestCommandPokedex(t *testing.T) {
	// Test pokedex command functionality
}
