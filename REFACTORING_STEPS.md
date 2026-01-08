# gowrite Refactoring Steps - Quick Reference

This document provides a quick, actionable checklist for refactoring the gowrite codebase. Each step is designed to be completed independently with verification.

## Quick Status

- ✅ **Step 0**: Initial state management refactoring (COMPLETED)
- ✅ **Step 1**: Extract persistence package (COMPLETED)
- ✅ **Step 2**: Extract commands package (COMPLETED)
- 🔄 **Step 3**: Extract analysis package (READY TO START)

---

## Step 3: Extract Analysis Package

**Time Estimate:** 30 minutes  
**Risk Level:** Low  
**Lines to Move:** ~150

### Files to Create
```bash
mkdir -p analysis
touch analysis/analysis.go
touch analysis/analysis_test.go
```

### Move from gowrite.go to analysis/analysis.go
1. Function: `CalculateReadability(text string) string`
2. Function: `AnalyzeTextForHemingway(text string) string`

### Move from gowrite_test.go to analysis/analysis_test.go
1. `TestCalculateReadability`
2. `TestAnalyzeTextForHemingway`
3. `TestAnalyzeTextForHemingway_MultipleIssues`
4. `BenchmarkCalculateReadability`
5. `BenchmarkAnalyzeTextForHemingway`

### Update gowrite.go
```go
import (
    "gowrite/analysis"
    // ... other imports
)

// Replace calls:
// CalculateReadability(text) → analysis.CalculateReadability(text)
// AnalyzeTextForHemingway(text) → analysis.AnalyzeTextForHemingway(text)
```

### Verification
```bash
go test ./...                    # All tests pass
go build -o gowrite gowrite.go   # Builds successfully
./gowrite                        # Manual test: run analyze command
```

---

## Step 4: Extract Spell Check Package

**Time Estimate:** 45 minutes  
**Risk Level:** Low  
**Lines to Move:** ~70

### Files to Create
```bash
mkdir -p spellcheck
touch spellcheck/spellcheck.go
touch spellcheck/spellcheck_test.go
```

### Functions to Extract

#### spellcheck/spellcheck.go
```go
package spellcheck

type Checker struct {
    dictionary map[string]bool
}

func NewChecker() *Checker
func (c *Checker) LoadDictionary(path string) error
func (c *Checker) CheckText(text string) []string
func (c *Checker) IsKnownWord(word string) bool
```

### Move from gowrite.go
1. Dictionary loading logic from `loadDictionary()`
2. Spell checking logic from `runSpellCheck()`
3. Global `dictionary` map → Checker struct

### Update gowrite.go
```go
import "gowrite/spellcheck"

// Create checker instance
checker := spellcheck.NewChecker()

// Replace loadDictionary() call
checker.LoadDictionary("dictionary.txt")

// Replace runSpellCheck() logic
unknowns := checker.CheckText(text)
```

### Verification
```bash
go test ./...
go build -o gowrite gowrite.go
# Test: Run spellcheck command in app
```

---

## Step 5: Extract UI Theme Package

**Time Estimate:** 1 hour  
**Risk Level:** Medium  
**Lines to Move:** ~100

### Files to Create
```bash
mkdir -p ui
touch ui/theme.go
touch ui/theme_test.go
```

### Functions to Extract

#### ui/theme.go
```go
package ui

type Theme struct {
    Name string
    // Color configurations
}

func ApplyTheme(name string, app *tview.Application, components ...tview.Primitive) error
func GetTheme(name string) (*Theme, error)
func ListThemes() []string
```

### Move from gowrite.go
1. `applyTheme()` function
2. Theme-related color configurations

### Benefits
- Easier to add new themes
- Theme logic testable in isolation
- Can load themes from config files

### Verification
```bash
go test ./...
go build -o gowrite gowrite.go
# Test: Run `theme light`, `theme dark`, `theme retro`
```

---

## Step 6: Extract Modal/Dialog Package

**Time Estimate:** 1 hour  
**Risk Level:** Medium  
**Lines to Move:** ~150

### Files to Create
```bash
touch ui/modals.go
touch ui/modals_test.go
```

### Functions to Extract

#### ui/modals.go
```go
package ui

type ModalBuilder struct {
    app   *tview.Application
    pages *tview.Pages
}

func NewModalBuilder(app *tview.Application, pages *tview.Pages) *ModalBuilder
func (m *ModalBuilder) ShowInfo(title, text string, onClose func())
func (m *ModalBuilder) ShowConfirm(title, text string, onYes, onNo func())
func (m *ModalBuilder) ShowFilePicker(files []string, onSelect func(string))
```

### Move from gowrite.go
1. `showModal()` → `ModalBuilder.ShowInfo()`
2. `showYesNoModal()` → `ModalBuilder.ShowConfirm()`
3. `showFilePicker()` → `ModalBuilder.ShowFilePicker()`

### Verification
```bash
go test ./...
go build -o gowrite gowrite.go
# Test: Trigger various modals (save, delete, etc.)
```

---

## Step 7: Extract View Management

**Time Estimate:** 1.5 hours  
**Risk Level:** High  
**Lines to Move:** ~200

### Files to Create
```bash
touch ui/views.go
touch ui/views_test.go
```

### Functions to Extract

#### ui/views.go
```go
package ui

type ViewController struct {
    app       *tview.Application
    mainView  *tview.Grid
    pages     *tview.Pages
    state     *state.AppState
    // UI components
}

func NewViewController(app *tview.Application, state *state.AppState) *ViewController
func (v *ViewController) SetView(viewType int)
func (v *ViewController) ToggleNotes()
func (v *ViewController) ToggleWiki()
func (v *ViewController) ToggleFocus()
```

### Move from gowrite.go
1. `setView()` function and all view transition logic
2. `toggleNotes()`, `toggleWiki()`, `toggleFocus()`
3. View layout configuration

### Verification
```bash
go test ./...
go build -o gowrite gowrite.go
# Test: Switch between all views (Ctrl-N, Ctrl-W, analyze, etc.)
```

---

## Step 8: Extract Command Extensions

**Time Estimate:** 2 hours  
**Risk Level:** Medium  
**Lines to Move:** ~200

### Update commands/commands.go

Add new command handlers:
```go
func init() {
    // Existing
    Registry["save"] = saveHandler
    Registry["open"] = openHandler
    Registry["load"] = openHandler
    Registry["export"] = exportHandler
    Registry["structure"] = structureHandler
    
    // New
    Registry["wordcount"] = wordcountHandler
    Registry["search"] = searchHandler
    Registry["theme"] = themeHandler
    Registry["analyze"] = analyzeHandler
    Registry["spellcheck"] = spellcheckHandler
    Registry["import"] = importHandler
    Registry["chapter"] = chapterHandler
    Registry["wiki"] = wikiHandler
}
```

### Move from gowrite.go handleCommand()
Extract switch cases into individual command handlers

### Verification
```bash
go test ./...
go build -o gowrite gowrite.go
# Test: Run all commands from command palette
```

---

## Step 9: Extract Chapter/Wiki Operations

**Time Estimate:** 1 hour  
**Risk Level:** Low  
**Lines to Move:** ~150

### Files to Create
```bash
touch ui/chapter_ops.go
touch ui/wiki_ops.go
```

### Functions to Extract

#### ui/chapter_ops.go
```go
package ui

func LoadChapter(index int, state *state.AppState, textArea, notesArea *tview.TextArea)
func SaveCurrentChapter(state *state.AppState, textArea, notesArea *tview.TextArea)
func DeleteChapter(index int, state *state.AppState, modal ModalBuilder)
func RenameChapter(index int, name string, state *state.AppState, modal ModalBuilder)
```

#### ui/wiki_ops.go
```go
package ui

func LoadWiki(index int, state *state.AppState, wikiArea *tview.TextArea, wikiList *tview.List)
func SaveCurrentWiki(state *state.AppState, wikiArea *tview.TextArea)
func DeleteWiki(index int, state *state.AppState, modal ModalBuilder)
func RenameWiki(index int, name string, state *state.AppState)
```

### Verification
```bash
go test ./...
go build -o gowrite gowrite.go
# Test: Create, rename, delete chapters and wiki entries
```

---

## Step 10: Extract Input Handling

**Time Estimate:** 1.5 hours  
**Risk Level:** Medium  
**Lines to Move:** ~120

### Files to Create
```bash
touch ui/input.go
```

### Functions to Extract

#### ui/input.go
```go
package ui

type KeyHandler struct {
    app   *tview.Application
    state *state.AppState
}

func NewKeyHandler(app *tview.Application, state *state.AppState) *KeyHandler
func (k *KeyHandler) SetupGlobalKeys()
func (k *KeyHandler) SetupTextAreaKeys(area *tview.TextArea)
func (k *KeyHandler) SetupWikiKeys(list *tview.List, area *tview.TextArea)
func (k *KeyHandler) SetupCommandPaletteKeys(palette *tview.InputField)
```

### Move from gowrite.go
1. Global input capture logic
2. Widget-specific key bindings
3. Command palette handling

### Verification
```bash
go test ./...
go build -o gowrite gowrite.go
# Test: All keyboard shortcuts (Ctrl-E, Ctrl-N, Ctrl-W, F1, etc.)
```

---

## Testing Checklist (After Each Step)

### Automated Tests
- [ ] `go test ./...` - All packages
- [ ] `go test -v` - Verbose output
- [ ] `go test -race` - Race condition detection
- [ ] `go test -cover` - Coverage report
- [ ] `go test -bench=.` - Benchmarks

### Build Tests
- [ ] `go build -o gowrite gowrite.go` - Successful build
- [ ] `go vet ./...` - No vet warnings
- [ ] `golint ./...` - Lint check (if installed)

### Manual Tests
- [ ] Application starts without errors
- [ ] Can create and edit chapters
- [ ] Can save and load projects
- [ ] Can export to text
- [ ] Theme switching works
- [ ] Analysis mode works
- [ ] Spell checking works
- [ ] Wiki functionality works
- [ ] All keyboard shortcuts work
- [ ] Command palette works

### Regression Tests
- [ ] Existing projects can be loaded
- [ ] All structure templates work
- [ ] Import functionality works
- [ ] Autosave works
- [ ] File picker works

---

## Common Issues and Solutions

### Issue: Import cycle
**Solution:** Ensure packages don't import each other. Use interfaces or move shared code to a common package.

### Issue: Tests fail after refactoring
**Solution:** Update test imports. Ensure test data is still accessible.

### Issue: Application doesn't start
**Solution:** Check for nil pointer dereferences. Ensure all components are properly initialized.

### Issue: Features don't work
**Solution:** Verify all function calls are updated. Check that callbacks and closures reference correct variables.

---

## Final Goal

```
gowrite/
├── analysis/           # Text analysis and readability
│   ├── analysis.go
│   └── analysis_test.go
├── commands/           # Command handlers
│   ├── commands.go
│   └── commands_test.go
├── persistence/        # File I/O
│   ├── persistence.go
│   └── persistence_test.go
├── spellcheck/         # Spell checking
│   ├── spellcheck.go
│   └── spellcheck_test.go
├── state/              # Application state
│   ├── state.go
│   └── state_test.go
├── ui/                 # UI components
│   ├── chapter_ops.go
│   ├── input.go
│   ├── modals.go
│   ├── theme.go
│   ├── views.go
│   └── wiki_ops.go
├── gowrite.go          # Main (reduced to ~300 lines)
├── gowrite_test.go     # Integration tests
└── dictionary.txt      # Dictionary data
```

**Final Line Count:**
- Before: `gowrite.go` = 1,427 lines
- After: `gowrite.go` = ~300 lines (79% reduction)

---

## Next Steps

1. Start with **Step 3: Extract Analysis Package** (lowest risk)
2. Proceed sequentially through steps
3. Commit after each completed step
4. Update this document with notes and deviations
5. Create pull requests for review

**Remember:** Make small, incremental changes. Test thoroughly. Document as you go.

---

Last Updated: 2026-01-08
