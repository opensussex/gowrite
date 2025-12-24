package main

import (
	"strings"
	"testing"
)

func TestCalculateReadability(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "simple sentence",
			text:     "The cat sat on the mat.",
			expected: "Reading Age: 5-6 (Grade 1)",
		},
		{
			name:     "complex text",
			text:     "The quick brown fox jumps over the lazy dog. This is a test sentence with more complexity and additional words to increase the grade level.",
			expected: "Reading Age: 11-12 (Grade 7)",
		},
		{
			name:     "empty text",
			text:     "",
			expected: "Reading Age: 5-6 (Grade 1)",
		},
		{
			name:     "single word",
			text:     "Hello",
			expected: "Reading Age: 7-8 (Grade 3)",
		},
		{
			name:     "adult level text",
			text:     "The implementation of sophisticated algorithmic methodologies necessitates comprehensive understanding of computational complexity theory and its practical applications in contemporary software engineering paradigms.",
			expected: "Reading Age: 18+ (Adult) (Grade 32)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateReadability(tt.text)
			if result != tt.expected {
				t.Errorf("CalculateReadability() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestAnalyzeTextForHemingway(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		contains []string
	}{
		{
			name:     "detects adverbs",
			text:     "He ran quickly.",
			contains: []string{"[blue]", "quickly"},
		},
		{
			name:     "detects passive voice",
			text:     "The ball was kicked.",
			contains: []string{"[green]", "was"},
		},
		{
			name:     "detects long sentences",
			text:     "This is a very long sentence with more than fourteen words in it to trigger the yellow warning for readability.",
			contains: []string{"[yellow]"},
		},
		{
			name:     "detects very long sentences",
			text:     "This is an extremely long sentence with more than twenty words in it which should trigger the red warning for very hard readability issues that need attention.",
			contains: []string{"[red]"},
		},
		{
			name:     "empty text",
			text:     "",
			contains: []string{},
		},
		{
			name:     "normal text no issues",
			text:     "The cat sat on the mat.",
			contains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AnalyzeTextForHemingway(tt.text)
			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("AnalyzeTextForHemingway() result missing %q\nGot: %s", substr, result)
				}
			}
		})
	}
}

func TestAnalyzeTextForHemingway_MultipleIssues(t *testing.T) {
	text := "He ran quickly and was kicked easily by the guards."
	result := AnalyzeTextForHemingway(text)

	// Should detect adverbs
	if !strings.Contains(result, "[blue]") {
		t.Error("Expected to detect adverbs (blue)")
	}

	// Should detect passive voice
	if !strings.Contains(result, "[green]") {
		t.Error("Expected to detect passive voice (green)")
	}
}

func TestViewConstants(t *testing.T) {
	// Verify constants are defined and unique
	views := []int{ViewMain, ViewNotes, ViewAnalyze, ViewWiki}
	seen := make(map[int]bool)

	for _, v := range views {
		if seen[v] {
			t.Errorf("Duplicate view constant value: %d", v)
		}
		seen[v] = true
	}

	// Verify they start at 0 and increment
	if ViewMain != 0 {
		t.Errorf("ViewMain should be 0, got %d", ViewMain)
	}
	if ViewNotes != 1 {
		t.Errorf("ViewNotes should be 1, got %d", ViewNotes)
	}
	if ViewAnalyze != 2 {
		t.Errorf("ViewAnalyze should be 2, got %d", ViewAnalyze)
	}
	if ViewWiki != 3 {
		t.Errorf("ViewWiki should be 3, got %d", ViewWiki)
	}
}

func TestTargetWidth(t *testing.T) {
	if TargetWidth != 85 {
		t.Errorf("TargetWidth should be 85, got %d", TargetWidth)
	}
}

// Benchmark tests
func BenchmarkCalculateReadability(b *testing.B) {
	text := strings.Repeat("This is a test sentence. ", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateReadability(text)
	}
}

func BenchmarkAnalyzeTextForHemingway(b *testing.B) {
	text := strings.Repeat("He ran quickly through the forest. The trees were swaying. ", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AnalyzeTextForHemingway(text)
	}
}
