# gowrite Refactoring Guide

## Overview

This document outlines a comprehensive refactoring plan for the gowrite terminal-based writing tool. The goal is to improve code organization, maintainability, and testability while preserving all existing functionality.

## Current State

### Existing Structure (✅ Completed)

The codebase has already undergone initial refactoring:

1. **state/** - Application state management
   - `AppState` - Central state container with mutex protection
   - Chapter and Wiki management
   - View state tracking

2. **persistence/** - File I/O operations
   - `SaveProject()` - JSON persistence
   - `LoadProject()` - JSON loading with backward compatibility
   - `ExportText()` - Plain text export

3. **commands/** - Command registry and handlers
   - Command handler interface
   - Built-in commands: save, open, load, export, structure

4. **gowrite.go** - Main application (1,427 lines)
   - UI setup and management
   - Event handling
   - Theme management
   - Modal dialogs
   - Analysis functions
   - Spell checking
   - Command processing

### Test Coverage

- ✅ `gowrite_test.go` - Tests for readability and analysis functions
- ✅ Benchmark tests included
- ✅ All tests passing

## Refactoring Steps

### Step 1: Extract UI Package ⭐ High Priority

**Goal:** Separate UI component creation and management from application logic.

**Files to Create:**
- `ui/ui.go` - Main UI structures and setup
- `ui/theme.go` - Theme management
- `ui/components.go` - Reusable UI components

**Functions to Move:**
```go
// From gowrite.go to ui/theme.go
- applyTheme()

// From gowrite.go to ui/components.go
- setupTextArea() (new function to extract textArea setup)
- setupNotesArea() (new function to extract notesArea setup)
- setupWikiComponents() (new function to extract wiki UI setup)
- setupCommandPalette() (new function to extract commandPalette setup)
```

**Benefits:**
- Cleaner separation of concerns
- Easier theme customization
- Reusable UI components
- Better testability

**Estimated Impact:** ~200 lines moved from gowrite.go

---

### Step 2: Extract Modal/Dialog Package

**Goal:** Centralize modal dialog creation and management.

**Files to Create:**
- `ui/modals.go` - Modal dialog builders

**Functions to Move:**
```go
// From gowrite.go to ui/modals.go
- showModal()
- showYesNoModal()
- showFilePicker() (file picker modal)
```

**New Functions to Create:**
```go
- NewInfoModal(title, text string) - Simple info modal
- NewConfirmModal(title, text string, onConfirm func()) - Yes/No modal
- NewFilePickerModal(files []string, onSelect func(string)) - File picker
```

**Benefits:**
- Consistent modal behavior
- Easier to add new modal types
- Better error handling
- Testable modal logic

**Estimated Impact:** ~150 lines moved from gowrite.go

---

### Step 3: Extract Spell Check Package

**Goal:** Move spell checking logic to dedicated package.

**Files to Create:**
- `spellcheck/spellcheck.go` - Spell checking functionality
- `spellcheck/spellcheck_test.go` - Tests for spell checking

**Functions to Move:**
```go
// From gowrite.go to spellcheck/spellcheck.go
- loadDictionary() → LoadDictionary(path string) (map[string]bool, error)
- runSpellCheck() → CheckText(text string, dict map[string]bool) []string

// New helper functions
- IsWord(s string) bool - Check if string is valid word
- NormalizeWord(s string) string - Clean and normalize word
- LoadDictionaryOnce() - Singleton pattern for dictionary loading
```

**Benefits:**
- Isolated spell checking logic
- Testable spell checking
- Can add more sophisticated spell checking
- Dictionary can be reused across components

**Estimated Impact:** ~70 lines moved from gowrite.go

---

### Step 4: Extract Analysis Package

**Goal:** Move readability and Hemingway analysis to dedicated package.

**Files to Create:**
- `analysis/analysis.go` - Analysis functions
- `analysis/analysis_test.go` - Move existing tests here

**Functions to Move:**
```go
// From gowrite.go to analysis/analysis.go
- CalculateReadability() (already extracted, just move to package)
- AnalyzeTextForHemingway() (already extracted, just move to package)

// Move tests from gowrite_test.go to analysis/analysis_test.go
- TestCalculateReadability
- TestAnalyzeTextForHemingway
- TestAnalyzeTextForHemingway_MultipleIssues
- BenchmarkCalculateReadability
- BenchmarkAnalyzeTextForHemingway
```

**New Functions to Add:**
```go
- CountWords(text string) int - Word counting utility
- CountSentences(text string) int - Sentence counting utility
- GetReadingLevel(ari float64) string - Convert ARI to reading level
```

**Benefits:**
- All analysis logic in one place
- Easier to add new analysis types
- Better organized tests
- Can be used by other tools

**Estimated Impact:** ~150 lines moved from gowrite.go

---

### Step 5: Extract View Management Package

**Goal:** Separate view state management and transitions.

**Files to Create:**
- `ui/views.go` - View management and transitions

**Functions to Move:**
```go
// From gowrite.go to ui/views.go
- setView()
- toggleNotes()
- toggleWiki()
- toggleFocus()

// New functions to create
- type ViewController interface - View management contract
- SetupMainView() - Configure main editor view
- SetupNotesView() - Configure notes view
- SetupAnalysisView() - Configure analysis view
- SetupWikiView() - Configure wiki view
```

**Benefits:**
- Centralized view management
- Easier to add new views
- Better state transitions
- Testable view logic

**Estimated Impact:** ~120 lines moved from gowrite.go

---

### Step 6: Extract Chapter/Wiki Operations

**Goal:** Move chapter and wiki UI operations to helper package.

**Files to Create:**
- `ui/chapter_ops.go` - Chapter UI operations
- `ui/wiki_ops.go` - Wiki UI operations

**Functions to Move:**
```go
// From gowrite.go to ui/chapter_ops.go
- loadChapter()
- deleteChapter()
- renameChapter()
- saveCurrentChapter() (wrap state.AppState method)

// From gowrite.go to ui/wiki_ops.go
- loadWiki()
- deleteWiki()
- renameWiki()
- saveCurrentWiki() (wrap state.AppState method)
```

**Benefits:**
- Isolated chapter/wiki operations
- Easier to maintain
- Clearer separation between state and UI
- Can add bulk operations easily

**Estimated Impact:** ~100 lines moved from gowrite.go

---

### Step 7: Extract Command Handler Extensions

**Goal:** Move remaining command logic to commands package.

**Files to Update:**
- `commands/commands.go` - Add more command handlers

**New Commands to Add:**
```go
// Add to commands package
- wordcountHandler() - Word count command
- searchHandler() - Search command
- themeHandler() - Theme switching command
- analyzeHandler() - Analysis command
- spellcheckHandler() - Spell check command
- importHandler() - Import command
- chapterHandler() - Chapter management commands
- wikiHandler() - Wiki management commands
```

**Benefits:**
- All command logic in commands package
- Consistent command interface
- Easier to add new commands
- Better testability

**Estimated Impact:** ~200 lines moved from gowrite.go

---

### Step 8: Extract Input Handling

**Goal:** Centralize keyboard input handling.

**Files to Create:**
- `ui/input.go` - Input handling and key bindings

**Functions to Create:**
```go
// New file: ui/input.go
- SetupGlobalKeys(app *tview.Application) - Global key bindings
- SetupTextAreaKeys(area *tview.TextArea) - Text area key bindings
- SetupWikiKeys(list *tview.List, area *tview.TextArea) - Wiki key bindings
- HandleCommandInput(cmd string) - Command palette handler
```

**Benefits:**
- Centralized key binding management
- Easier to customize shortcuts
- Better documentation of shortcuts
- Can add key binding help screen

**Estimated Impact:** ~100 lines moved from gowrite.go

---

## Refactoring Best Practices

### 1. Extract in Small Steps
- Make one refactoring at a time
- Run tests after each change
- Commit after each successful refactoring

### 2. Maintain Backward Compatibility
- Keep existing functionality unchanged
- Preserve all command interfaces
- Maintain file format compatibility

### 3. Write Tests First
- Add tests for extracted functions
- Ensure test coverage doesn't decrease
- Add integration tests where needed

### 4. Document Interfaces
- Add godoc comments to all exported functions
- Document package purpose
- Add usage examples

### 5. Use Constructor Functions
- Create `New*()` functions for complex types
- Initialize with sensible defaults
- Make zero values useful where possible

## Testing Strategy

### Unit Tests
Each new package should have comprehensive unit tests:
- `analysis/analysis_test.go` - Analysis functions
- `spellcheck/spellcheck_test.go` - Spell checking
- `ui/theme_test.go` - Theme application
- `ui/modals_test.go` - Modal creation

### Integration Tests
Add integration tests for:
- View transitions
- Command processing
- File operations with state

### Benchmark Tests
Maintain benchmark tests for:
- Analysis functions (already exists)
- Spell checking
- Large text operations

## Migration Path

### Phase 1: Low-Risk Extractions (Weeks 1-2)
1. ✅ Extract state package (DONE)
2. ✅ Extract persistence package (DONE)
3. ✅ Extract commands package (DONE)
4. Extract analysis package (Step 4)
5. Extract spell check package (Step 3)

### Phase 2: UI Refactoring (Weeks 3-4)
1. Extract modal/dialog package (Step 2)
2. Extract theme management (Step 1)
3. Extract view management (Step 5)

### Phase 3: Advanced Refactoring (Weeks 5-6)
1. Extract chapter/wiki operations (Step 6)
2. Extract input handling (Step 8)
3. Complete command handler extraction (Step 7)

### Phase 4: Polish and Optimization (Week 7)
1. Add comprehensive tests
2. Update documentation
3. Performance optimization
4. Code cleanup

## Expected Outcomes

### Before Refactoring
- `gowrite.go`: 1,427 lines
- Hard to test UI logic
- Mixed concerns
- Difficult to extend

### After Refactoring
- `gowrite.go`: ~300 lines (main setup and wiring)
- `analysis/`: Readability and Hemingway analysis
- `spellcheck/`: Spell checking logic
- `ui/`: UI components, themes, modals, views, input
- `commands/`: All command handlers
- `state/`: Application state (existing)
- `persistence/`: File I/O (existing)

### Benefits
- ✅ Better code organization
- ✅ Improved testability
- ✅ Easier maintenance
- ✅ Simpler extension
- ✅ Better documentation
- ✅ Faster development

## Quick Start Guide

To begin refactoring, start with Step 4 (Extract Analysis Package) as it's the lowest risk:

```bash
# 1. Create the analysis package
mkdir analysis
touch analysis/analysis.go
touch analysis/analysis_test.go

# 2. Move functions from gowrite.go to analysis/analysis.go
# Move: CalculateReadability, AnalyzeTextForHemingway

# 3. Move tests from gowrite_test.go to analysis/analysis_test.go
# Move: All analysis-related tests

# 4. Update imports in gowrite.go
# Add: "gowrite/analysis"

# 5. Run tests
go test ./...

# 6. Build and verify
go build -o gowrite gowrite.go
./gowrite  # Manual verification
```

## Verification

After each refactoring step:

1. **Run tests:** `go test -v ./...`
2. **Run benchmarks:** `go test -bench=. -benchmem`
3. **Check coverage:** `go test -cover ./...`
4. **Build application:** `go build -o gowrite gowrite.go`
5. **Manual testing:** Run the application and test affected features
6. **Run verification script:** `./verify_refactoring.sh`

## Notes

- The refactoring should be done incrementally
- Each step should be a separate commit
- All tests must pass before proceeding to next step
- Update this document as refactoring progresses
- Document any deviations from the plan

## Resources

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)

---

Last Updated: 2026-01-08
Status: Planning Phase
