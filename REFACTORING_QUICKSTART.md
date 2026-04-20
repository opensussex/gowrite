# Quick Start: Refactoring gowrite

This guide helps you start refactoring immediately. Pick any step and follow the instructions.

## Prerequisites

```bash
# Verify your environment
go version  # Should be 1.18+
go test ./...  # All tests should pass
go build -o gowrite gowrite.go  # Should build successfully
```

## Current Status (Run this anytime)

```bash
./verify_refactoring.sh
```

This will show:
- ✅ Completed refactoring steps
- ⏸️  Pending refactoring steps
- Current progress percentage
- Next recommended step

---

## 🚀 START HERE: Step 3 - Extract Analysis Package

**This is the easiest and lowest risk refactoring to start with.**

### 1. Create the package structure

```bash
mkdir -p analysis
```

### 2. Create analysis/analysis.go

```bash
cat > analysis/analysis.go << 'EOF'
package analysis

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"unicode"
)

// CalculateReadability computes ARI grade level and returns age range
func CalculateReadability(text string) string {
	words := len(strings.Fields(text))
	sentences := strings.Count(text, ".") + strings.Count(text, "!") + strings.Count(text, "?")
	if sentences == 0 {
		sentences = 1
	}

	chars := 0
	for _, r := range text {
		if !unicode.IsSpace(r) {
			chars++
		}
	}
	if words == 0 {
		words = 1
	}

	ari := 4.71*(float64(chars)/float64(words)) + 0.5*(float64(words)/float64(sentences)) - 21.43
	grade := int(math.Ceil(ari))
	if grade < 1 {
		grade = 1
	}

	ageRange := "Adult"
	switch grade {
	case 1:
		ageRange = "5-6"
	case 2:
		ageRange = "6-7"
	case 3:
		ageRange = "7-8"
	case 4:
		ageRange = "8-9"
	case 5:
		ageRange = "9-10"
	case 6:
		ageRange = "10-11"
	case 7:
		ageRange = "11-12"
	case 8:
		ageRange = "12-13"
	case 9:
		ageRange = "13-14"
	case 10:
		ageRange = "14-15"
	case 11:
		ageRange = "15-16"
	case 12:
		ageRange = "16-17"
	case 13:
		ageRange = "17-18"
	default:
		ageRange = "18+ (Adult)"
	}

	return fmt.Sprintf("Reading Age: %s (Grade %d)", ageRange, grade)
}

// AnalyzeTextForHemingway returns text with color markup for prose issues
func AnalyzeTextForHemingway(text string) string {
	adverbRegex := regexp.MustCompile(`(?i)\b(\w+ly)\b`)
	passiveRegex := regexp.MustCompile(`(?i)\b(am|are|is|was|were|be|been|being)\b\s+(\w+ed)\b`)

	paragraphs := strings.Split(text, "\n")
	var processedText strings.Builder

	for _, para := range paragraphs {
		if strings.TrimSpace(para) == "" {
			processedText.WriteString("\n")
			continue
		}

		sentenceRe := regexp.MustCompile(`[^.!?]+[.!?]*`)
		matches := sentenceRe.FindAllString(para, -1)

		for _, s := range matches {
			wordCount := len(strings.Fields(s))
			coloredS := s

			prefix := ""
			suffix := ""

			if wordCount > 20 {
				prefix = "[red]"
				suffix = "[-]"
			} else if wordCount > 14 {
				prefix = "[yellow]"
				suffix = "[-]"
			}

			coloredS = adverbRegex.ReplaceAllStringFunc(coloredS, func(m string) string {
				return "[blue]" + m + "[-]" + prefix
			})

			coloredS = passiveRegex.ReplaceAllStringFunc(coloredS, func(m string) string {
				return "[green]" + m + "[-]" + prefix
			})

			processedText.WriteString(prefix + coloredS + suffix + " ")
		}
		processedText.WriteString("\n")
	}

	return processedText.String()
}
EOF
```

### 3. Create analysis/analysis_test.go

```bash
cat > analysis/analysis_test.go << 'EOF'
package analysis

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
EOF
```

### 4. Update gowrite.go

```bash
# Add import at the top
# Find: import (
# After it add: "gowrite/analysis"

# Replace function calls:
# Find: CalculateReadability(
# Replace with: analysis.CalculateReadability(

# Find: AnalyzeTextForHemingway(
# Replace with: analysis.AnalyzeTextForHemingway(

# Remove the original function definitions from gowrite.go
# (Lines ~40-145)
```

### 5. Update gowrite_test.go

```bash
# Remove the analysis tests from gowrite_test.go:
# - TestCalculateReadability
# - TestAnalyzeTextForHemingway
# - TestAnalyzeTextForHemingway_MultipleIssues
# - BenchmarkCalculateReadability
# - BenchmarkAnalyzeTextForHemingway

# Keep only:
# - TestViewConstants
# - TestTargetWidth
```

### 6. Test everything

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./analysis -v

# Run benchmarks
go test ./analysis -bench=. -benchmem

# Build the application
go build -o gowrite gowrite.go

# Verify refactoring progress
./verify_refactoring.sh
```

### 7. Manual verification

```bash
# Run the application
./gowrite

# Test the analyze command:
# 1. Type some text
# 2. Press Ctrl-E
# 3. Type: analyze
# 4. Verify the Hemingway analysis appears with colored text
# 5. Press Esc to return
```

### 8. Commit your changes

```bash
git add .
git commit -m "Extract analysis package (Step 3)

- Move CalculateReadability to analysis package
- Move AnalyzeTextForHemingway to analysis package
- Move all analysis tests to analysis package
- Update imports in gowrite.go
- All tests passing
"
```

---

## ✅ Verification Checklist

After completing Step 3, verify:

- [ ] `analysis/` directory exists
- [ ] `analysis/analysis.go` contains both functions
- [ ] `analysis/analysis_test.go` contains all tests
- [ ] `gowrite.go` imports `"gowrite/analysis"`
- [ ] `gowrite.go` calls `analysis.CalculateReadability()`
- [ ] `gowrite.go` calls `analysis.AnalyzeTextForHemingway()`
- [ ] Original functions removed from `gowrite.go`
- [ ] Tests removed from `gowrite_test.go`
- [ ] `go test ./...` passes
- [ ] `go build -o gowrite gowrite.go` succeeds
- [ ] Application runs and analyze command works
- [ ] `./verify_refactoring.sh` shows 4/10 steps (40%)

---

## 🎯 What's Next?

After Step 3, you can choose:

### Option A: Continue Incrementally (Recommended)
Go to **Step 4: Extract Spell Check Package** (see REFACTORING_STEPS.md)

### Option B: Focus on UI
Jump to **Step 5: Extract UI Theme Package** (see REFACTORING_STEPS.md)

### Option C: Add More Commands
Jump to **Step 8: Extract Command Extensions** (see REFACTORING_STEPS.md)

---

## 🆘 Troubleshooting

### Tests fail with "undefined: CalculateReadability"
**Fix:** You haven't updated the import in gowrite.go. Add `"gowrite/analysis"` to imports.

### Tests fail with "undefined: analysis"
**Fix:** The import path might be wrong. Verify your module path in `go.mod`.

### Build fails with import cycle
**Fix:** Make sure analysis package doesn't import gowrite package. It should be standalone.

### Application crashes when running analyze
**Fix:** Verify you replaced ALL calls to `CalculateReadability` and `AnalyzeTextForHemingway` with the `analysis.` prefix.

### Verification script shows 30% instead of 40%
**Fix:** The analysis package might not be detected. Ensure `analysis/analysis.go` exists and contains the functions.

---

## 📚 Additional Resources

- [Full refactoring guide](REFACTORING.md)
- [Detailed step-by-step instructions](REFACTORING_STEPS.md)
- [Verification script](verify_refactoring.sh)

---

## 💡 Pro Tips

1. **Work in small commits**: Commit after each function is moved and tested
2. **Run tests frequently**: `go test ./...` after every change
3. **Use the verification script**: `./verify_refactoring.sh` to track progress
4. **Manual test critical paths**: Don't rely only on unit tests
5. **Keep the app buildable**: Never break the build for more than 5 minutes

---

**Time Estimate for Step 3:** 30-45 minutes including testing

**Ready? Let's refactor! 🚀**
