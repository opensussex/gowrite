// ...existing code...
package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"gowrite/commands"
	"gowrite/persistence"
	"gowrite/state"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Aliases to shared state types
type Chapter = state.Chapter
type WikiEntry = state.WikiEntry
type Project = state.Project

// View state constants
const (
	ViewMain = iota
	ViewNotes
	ViewAnalyze
	ViewWiki
)

// TargetWidth is the centered view column width
const TargetWidth = 85

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

func main() {
	// --- 0. THEME SETUP ---
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorBlack
	tview.Styles.ContrastBackgroundColor = tcell.ColorDarkBlue
	tview.Styles.MoreContrastBackgroundColor = tcell.ColorGreen
	tview.Styles.BorderColor = tcell.ColorDarkGray
	tview.Styles.TitleColor = tcell.ColorYellow
	tview.Styles.GraphicsColor = tcell.ColorYellow
	tview.Styles.PrimaryTextColor = tcell.ColorWhite
	tview.Styles.SecondaryTextColor = tcell.ColorYellow
	tview.Styles.TertiaryTextColor = tcell.ColorGreen
	tview.Styles.InverseTextColor = tcell.ColorBlue
	tview.Styles.ContrastSecondaryTextColor = tcell.ColorDarkCyan

	app := tview.NewApplication()

	// --- 1. Data Management ---
	appState := state.NewAppState()
	appState.SetView(ViewMain)

	// Visual state tracked in AppState

	dictionary := make(map[string]bool)
	dictionaryLoaded := false

	// --- 2. Setup Main Components ---

	// MAIN EDITOR
	initialCh := appState.Snapshot().Chapters[0]
	textArea := tview.NewTextArea()
	textArea.SetWrap(true)
	textArea.SetPlaceholder("Start writing your masterpiece...")
	textArea.SetTextStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite))
	textArea.SetTitle(fmt.Sprintf("gowrite - Chapter 1: %s", initialCh.Title))
	textArea.SetBorder(true)
	textArea.SetBorderPadding(1, 1, 2, 2)

	// NOTES EDITOR
	notesArea := tview.NewTextArea()
	notesArea.SetWrap(true)
	notesArea.SetPlaceholder("Scene ideas, plot points, and reminders...")
	notesArea.SetTextStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorYellow))
	notesArea.SetTitle("SCENE NOTES")
	notesArea.SetBorder(true)
	notesArea.SetBorderPadding(1, 1, 2, 2)

	// WIKI LIST (Story Wiki)
	wikiList := tview.NewList()
	wikiList.ShowSecondaryText(false)
	wikiList.SetBorder(true)
	wikiList.SetTitle("Story Wiki (Ctrl-W to Close)")
	wikiList.SetSelectedBackgroundColor(tview.Styles.TitleColor)
	wikiList.SetSelectedTextColor(tview.Styles.PrimitiveBackgroundColor)

	// WIKI TEXT AREA
	wikiArea := tview.NewTextArea()
	wikiArea.SetWrap(true)
	wikiArea.SetPlaceholder("Enter details for this entry...")
	wikiArea.SetTextStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorDarkCyan))
	wikiArea.SetTitle("Entry Content")
	wikiArea.SetBorder(true)
	wikiArea.SetBorderPadding(1, 1, 2, 2)

	// ANALYSIS VIEWER (Read Only)
	analysisView := tview.NewTextView()
	analysisView.SetDynamicColors(true)
	analysisView.SetWrap(true)
	analysisView.SetWordWrap(true)
	analysisView.SetTitle("HEMINGWAY ANALYSIS MODE")
	analysisView.SetBorder(true)
	analysisView.SetBorderPadding(1, 1, 2, 2)

	commandPalette := tview.NewInputField()
	commandPalette.SetLabel(" > ")
	commandPalette.SetFieldBackgroundColor(tcell.ColorBlack)
	commandPalette.SetFieldTextColor(tcell.ColorWhite)
	commandPalette.SetLabelColor(tcell.ColorYellow)
	commandPalette.SetPlaceholder("Type 'help' for commands")
	commandPalette.SetBorder(true)
	commandPalette.SetBorderPadding(0, 0, 1, 1)
	commandPalette.SetTitle("Command Palette")

	defaultHelpText := " F1: Help | Ctrl-E: Command Palette"
	helpInfo := tview.NewTextView()
	helpInfo.SetText(defaultHelpText)
	helpInfo.SetTextColor(tcell.ColorDarkGray)

	position := tview.NewTextView()
	position.SetDynamicColors(true)
	position.SetTextAlign(tview.AlignRight)

	pages := tview.NewPages()

	// Layout Grid
	mainView := tview.NewGrid()
	mainView.SetRows(0, 3, 1)
	mainView.AddItem(textArea, 0, 0, 1, 2, 0, 0, true)
	mainView.AddItem(commandPalette, 1, 0, 1, 2, 0, 0, false)
	mainView.AddItem(helpInfo, 2, 0, 1, 1, 0, 0, false)
	mainView.AddItem(position, 2, 1, 1, 1, 0, 0, false)

	// --- 3. THEME LOGIC ---

	applyTheme := func(name string) {
		name = strings.ToLower(name)
		analysisView.SetBackgroundColor(tcell.ColorBlack)

		switch name {
		case "light":
			tview.Styles.PrimitiveBackgroundColor = tcell.ColorWhite
			tview.Styles.ContrastBackgroundColor = tcell.ColorLightGray
			tview.Styles.BorderColor = tcell.ColorBlack
			tview.Styles.TitleColor = tcell.ColorDarkBlue
			tview.Styles.PrimaryTextColor = tcell.ColorBlack
			tview.Styles.SecondaryTextColor = tcell.ColorDarkBlue

			style := tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack)
			textArea.SetTextStyle(style).SetBackgroundColor(tcell.ColorWhite)

			notesArea.SetTextStyle(tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorDarkBlue))
			notesArea.SetBackgroundColor(tcell.ColorWhite)

			wikiArea.SetTextStyle(tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorDarkCyan))
			wikiArea.SetBackgroundColor(tcell.ColorWhite)

			analysisView.SetBackgroundColor(tcell.ColorWhite)

			commandPalette.SetFieldBackgroundColor(tcell.ColorWhite).SetFieldTextColor(tcell.ColorBlack).SetBackgroundColor(tcell.ColorWhite)
			helpInfo.SetTextColor(tcell.ColorDarkGray).SetBackgroundColor(tcell.ColorWhite)
			position.SetBackgroundColor(tcell.ColorWhite)

		case "retro":
			tview.Styles.PrimitiveBackgroundColor = tcell.ColorBlack
			tview.Styles.BorderColor = tcell.ColorGreen
			tview.Styles.TitleColor = tcell.ColorGreen
			tview.Styles.PrimaryTextColor = tcell.ColorGreen

			style := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorGreen)
			textArea.SetTextStyle(style).SetBackgroundColor(tcell.ColorBlack)

			notesArea.SetTextStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorDarkGreen))
			notesArea.SetBackgroundColor(tcell.ColorBlack)

			wikiArea.SetTextStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorDarkCyan))
			wikiArea.SetBackgroundColor(tcell.ColorBlack)

			analysisView.SetBackgroundColor(tcell.ColorBlack)

			commandPalette.SetFieldBackgroundColor(tcell.ColorBlack).SetFieldTextColor(tcell.ColorGreen).SetBackgroundColor(tcell.ColorBlack)
			helpInfo.SetTextColor(tcell.ColorGreen).SetBackgroundColor(tcell.ColorBlack)
			position.SetBackgroundColor(tcell.ColorBlack)

		case "dark":
			tview.Styles.PrimitiveBackgroundColor = tcell.ColorBlack
			tview.Styles.BorderColor = tcell.ColorDarkGray
			tview.Styles.TitleColor = tcell.ColorYellow
			tview.Styles.PrimaryTextColor = tcell.ColorWhite

			style := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
			textArea.SetTextStyle(style).SetBackgroundColor(tcell.ColorBlack)

			notesArea.SetTextStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorYellow))
			notesArea.SetBackgroundColor(tcell.ColorBlack)

			wikiArea.SetTextStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorDarkCyan))
			wikiArea.SetBackgroundColor(tcell.ColorBlack)

			analysisView.SetBackgroundColor(tcell.ColorBlack)

			commandPalette.SetFieldBackgroundColor(tcell.ColorBlack).SetFieldTextColor(tcell.ColorWhite).SetBackgroundColor(tcell.ColorBlack)
			helpInfo.SetTextColor(tcell.ColorDarkGray).SetBackgroundColor(tcell.ColorBlack)
			position.SetBackgroundColor(tcell.ColorBlack)
		}
	}
	applyTheme("retro")

	// --- 4. Logic & Helper Functions ---

	// VIEW RESIZE LOGIC
	app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		w, _ := screen.Size()

		var hPadding int
		// If centered view is ON and screen is wide enough to justify it
		if appState.Centered() && w > TargetWidth+4 {
			hPadding = (w - TargetWidth) / 2
		} else {
			hPadding = 2 // Default small padding
		}

		// Apply to all text views
		textArea.SetBorderPadding(1, 1, hPadding, hPadding)
		notesArea.SetBorderPadding(1, 1, hPadding, hPadding)
		analysisView.SetBorderPadding(1, 1, hPadding, hPadding)
		wikiArea.SetBorderPadding(1, 1, 2, 2) // Wiki gets standard padding

		return false
	})

	// Forward-declared UI helpers
	var showModal func(title, text string)
	var showYesNoModal func(title, text string, onYes func())

	saveCurrentChapter := func() {
		if appState.Loading() {
			return
		}
		appState.SaveCurrentChapter(textArea.GetText(), notesArea.GetText())
	}

	saveCurrentWiki := func() {
		if appState.Loading() {
			return
		}
		appState.SaveCurrentWiki(wikiArea.GetText())
	}

	loadChapter := func(index int) {
		saveCurrentChapter()
		chapter, err := appState.LoadChapter(index)
		if err != nil {
			showModal("Error", err.Error())
			return
		}

		textArea.SetText(chapter.Content, false)
		notesArea.SetText(chapter.Notes, false)

		title := fmt.Sprintf("gowrite - Chapter %d: %s", index+1, chapter.Title)
		if appState.CurrentView() == ViewNotes {
			title += " (NOTES)"
		}
		textArea.SetTitle(title)
		notesArea.SetTitle(fmt.Sprintf("NOTES - Chapter %d", index+1))

		pages.HidePage("modal")

		if appState.CurrentView() == ViewNotes {
			app.SetFocus(notesArea)
		} else {
			app.SetFocus(textArea)
		}
	}

	// Forward declaration for recursion
	var loadWiki func(int)

	loadWiki = func(index int) {
		saveCurrentWiki()
		entry, err := appState.LoadWiki(index)
		if err != nil {
			showModal("Error", err.Error())
			return
		}

		wikiArea.SetText(entry.Content, false)
		wikiArea.SetTitle(fmt.Sprintf("Wiki: %s", entry.Title))

		wikiList.Clear()
		proj := appState.Snapshot()
		for i, w := range proj.Wiki {
			title := w.Title
			if i == appState.CurrentWikiIndex() {
				title += " *"
			}
			idx := i
			wikiList.AddItem(title, "", 0, func() {
				loadWiki(idx)
				app.SetFocus(wikiArea)
			})
		}
		wikiList.SetCurrentItem(appState.CurrentWikiIndex())
	}

	setView := func(viewType int) {
		if appState.CurrentView() == ViewWiki {
			saveCurrentWiki()
		} else {
			saveCurrentChapter()
		}

		appState.SetView(viewType)
		mainView.Clear()

		var activeWidget tview.Primitive
		var title string
		proj := appState.Snapshot()
		chapterIdx := appState.CurrentChapterIndex()
		if chapterIdx >= len(proj.Chapters) {
			chapterIdx = 0
		}
		chapter := proj.Chapters[chapterIdx]

		switch viewType {
		case ViewMain:
			activeWidget = textArea
			title = fmt.Sprintf("gowrite - Chapter %d: %s", chapterIdx+1, chapter.Title)
			helpInfo.SetText(defaultHelpText)
			mainView.SetColumns(0) // Reset to single column

		case ViewNotes:
			activeWidget = notesArea
			title = fmt.Sprintf("gowrite - Chapter %d: %s (NOTES)", chapterIdx+1, chapter.Title)
			helpInfo.SetText(" EDITING NOTES | Ctrl-N: Back | Ctrl-T: Center | Ctrl-F: Focus Mode")
			mainView.SetColumns(0) // Reset to single column

		case ViewAnalyze:
			activeWidget = analysisView
			title = "HEMINGWAY ANALYSIS MODE"
			helpInfo.SetText(" ANALYSIS | [Blue]Adverbs [Green]Passive [Yellow]Hard [Red]Very Hard | Esc: Exit")
			mainView.SetColumns(0) // Reset to single column

		case ViewWiki:
			// WIKI LAYOUT: List on left, Text on right
			activeWidget = wikiList
			title = "Story Wiki"
			helpInfo.SetText(" Wiki | Enter: Select | Tab: Edit Text | Ctrl-W: Close | 'wiki new/del' to manage")

			loadWiki(appState.CurrentWikiIndex())

			mainView.SetColumns(30, 0)
			mainView.SetRows(0, 3, 1)

			mainView.AddItem(wikiList, 0, 0, 1, 1, 0, 0, true)
			mainView.AddItem(wikiArea, 0, 1, 1, 1, 0, 0, false)
			mainView.AddItem(commandPalette, 1, 0, 1, 2, 0, 0, false)
			mainView.AddItem(helpInfo, 2, 0, 1, 1, 0, 0, false)
			mainView.AddItem(position, 2, 1, 1, 1, 0, 0, false)

			if appState.FocusMode() {
				mainView.SetRows(0)
				mainView.AddItem(wikiList, 0, 0, 1, 1, 0, 0, true)
				mainView.AddItem(wikiArea, 0, 1, 1, 1, 0, 0, false)
				wikiList.SetBorder(false)
				wikiArea.SetBorder(false)
			} else {
				wikiList.SetBorder(true)
				wikiArea.SetBorder(true)
			}

			app.SetFocus(wikiList)
			return // Exit function early, we handled the layout manually
		}

		// 3. Apply Layout for Standard Views (Main, Notes, Analyze)
		if appState.FocusMode() {
			// FOCUS: Single row, no borders, full height
			mainView.SetRows(0)
			mainView.AddItem(activeWidget, 0, 0, 1, 2, 0, 0, true)

			if v, ok := activeWidget.(*tview.TextArea); ok {
				v.SetBorder(false)
			}
			if v, ok := activeWidget.(*tview.TextView); ok {
				v.SetBorder(false)
			}
		} else {
			// NORMAL: 3 Rows, Borders on
			mainView.SetRows(0, 3, 1)
			mainView.AddItem(activeWidget, 0, 0, 1, 2, 0, 0, true)
			mainView.AddItem(commandPalette, 1, 0, 1, 2, 0, 0, false)
			mainView.AddItem(helpInfo, 2, 0, 1, 1, 0, 0, false)
			mainView.AddItem(position, 2, 1, 1, 1, 0, 0, false)

			if v, ok := activeWidget.(*tview.TextArea); ok {
				v.SetBorder(true).SetTitle(title)
			}
			if v, ok := activeWidget.(*tview.TextView); ok {
				v.SetBorder(true).SetTitle(title)
			}
		}

		// 4. Focus
		app.SetFocus(activeWidget)
	}

	toggleNotes := func() {
		view := appState.CurrentView()
		if view == ViewNotes || view == ViewWiki {
			setView(ViewMain)
		} else {
			setView(ViewNotes)
		}
	}

	toggleWiki := func() {
		if appState.CurrentView() == ViewWiki {
			setView(ViewMain)
		} else {
			setView(ViewWiki)
		}
	}

	toggleFocus := func() {
		appState.ToggleFocus()
		setView(appState.CurrentView())
	}

	showModal = func(title, text string) {
		modal := tview.NewModal()
		modal.SetText(text)
		modal.AddButtons([]string{"OK"})
		modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.HidePage("modal")
			// Restore focus
			switch appState.CurrentView() {
			case ViewNotes:
				app.SetFocus(notesArea)
			case ViewAnalyze:
				app.SetFocus(analysisView)
			case ViewWiki:
				app.SetFocus(wikiArea)
			default:
				app.SetFocus(textArea)
			}
		})

		modal.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
			if e.Key() == tcell.KeyEnter {
				pages.HidePage("modal")
				switch appState.CurrentView() {
				case ViewNotes:
					app.SetFocus(notesArea)
				case ViewWiki:
					app.SetFocus(wikiArea)
				default:
					app.SetFocus(textArea)
				}
				return nil
			}
			return e
		})

		modal.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
		modal.SetTextColor(tview.Styles.PrimaryTextColor)
		modal.SetButtonBackgroundColor(tview.Styles.TitleColor)
		modal.SetButtonTextColor(tview.Styles.PrimitiveBackgroundColor)
		pages.AddPage("modal", modal, true, true)
		app.SetFocus(modal)
	}

	showYesNoModal = func(title, text string, onYes func()) {
		modal := tview.NewModal()
		modal.SetText(text)
		modal.AddButtons([]string{"Yes", "No"})
		modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				onYes()
			}
			pages.HidePage("modal")
			switch appState.CurrentView() {
			case ViewNotes:
				app.SetFocus(notesArea)
			case ViewWiki:
				app.SetFocus(wikiArea)
			default:
				app.SetFocus(textArea)
			}
		})

		modal.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
		modal.SetTextColor(tview.Styles.PrimaryTextColor)
		modal.SetButtonBackgroundColor(tview.Styles.TitleColor)
		modal.SetButtonTextColor(tview.Styles.PrimitiveBackgroundColor)
		pages.AddPage("modal", modal, true, true)
		app.SetFocus(modal)
	}

	flashStatusMessage := func(msg string) {
		helpInfo.SetText(msg).SetTextColor(tcell.ColorGreen)
		go func() {
			time.Sleep(3 * time.Second)
			app.QueueUpdateDraw(func() {
				helpInfo.SetText(defaultHelpText).SetTextColor(tview.Styles.BorderColor)
			})
		}()
	}

	// --- CHAPTER OPS ---
	deleteChapter := func(index int) {
		proj := appState.Snapshot()
		if len(proj.Chapters) <= 1 {
			showModal("Error", "Cannot delete only chapter.")
			return
		}

		showYesNoModal("Confirm", fmt.Sprintf("Delete Chapter %d?", index+1), func() {
			newIdx, err := appState.DeleteChapter(index)
			if err != nil {
				showModal("Error", err.Error())
				return
			}
			loadChapter(newIdx)
		})
	}

	renameChapter := func(index int, newName string) {
		if err := appState.RenameChapter(index, newName); err != nil {
			showModal("Error", err.Error())
			return
		}
		if index == appState.CurrentChapterIndex() {
			loadChapter(index)
		} else {
			showModal("Success", fmt.Sprintf("Renamed Chapter %d to '%s'", index+1, newName))
		}
	}

	// --- STRUCTURE TEMPLATES ---
	applyStructure := func(name string) {
		if err := appState.ApplyStructure(name); err != nil {
			showModal("Error", err.Error())
			return
		}
		proj := appState.Snapshot()
		textArea.SetText(proj.Chapters[0].Content, false)
		notesArea.SetText(proj.Chapters[0].Notes, false)
		textArea.SetTitle(fmt.Sprintf("gowrite - Chapter 1: %s", proj.Chapters[0].Title))
		flashStatusMessage("Applied Structure: " + name)
	}

	// --- WIKI OPS ---
	deleteWiki := func(index int) {
		proj := appState.Snapshot()
		if len(proj.Wiki) <= 1 {
			showModal("Error", "Cannot delete the only wiki entry.")
			return
		}
		title := proj.Wiki[index].Title
		showYesNoModal("Confirm", fmt.Sprintf("Delete Wiki Entry '%s'?", title), func() {
			newIdx, err := appState.DeleteWiki(index)
			if err != nil {
				showModal("Error", err.Error())
				return
			}
			loadWiki(newIdx)
		})
	}

	renameWiki := func(index int, newName string) {
		if err := appState.RenameWiki(index, newName); err != nil {
			showModal("Error", err.Error())
			return
		}
		loadWiki(appState.CurrentWikiIndex())
	}

	// --- ANALYSIS LOGIC (Hemingway) ---

	runAnalysis := func() {
		text := textArea.GetText()
		processedText := AnalyzeTextForHemingway(text)
		analysisView.SetText(processedText)
		setView(ViewAnalyze)

		stats := CalculateReadability(text)
		key := "\n\n[::u]COLOR KEY[::-]\n" +
			"[blue]• Adverbs[-]\n" +
			"[green]• Passive Voice[-]\n" +
			"[yellow]• Hard Sentence (>14 words)[-]\n" +
			"[red]• Very Hard Sentence (>20 words)[-]"

		showModal("Readability Report", stats+key)
	}

	// --- SPELL CHECK ---
	loadDictionary := func() error {
		file, err := os.Open("dictionary.txt")
		if err != nil {
			return err
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			dictionary[strings.TrimSpace(strings.ToLower(scanner.Text()))] = true
		}
		dictionaryLoaded = true
		return scanner.Err()
	}

	runSpellCheck := func() {
		if !dictionaryLoaded {
			if err := loadDictionary(); err != nil {
				showModal("Error", "Could not load 'dictionary.txt'.")
				return
			}
		}

		targetArea := textArea
		view := appState.CurrentView()
		if view == ViewNotes {
			targetArea = notesArea
		} else if view == ViewWiki {
			targetArea = wikiArea
		}

		text := targetArea.GetText()
		words := strings.Fields(text)
		unknowns := make(map[string]bool)

		for _, rawWord := range words {
			cleanWord := strings.TrimFunc(rawWord, func(r rune) bool {
				return !unicode.IsLetter(r) && !unicode.IsNumber(r)
			})
			cleanWord = strings.ToLower(cleanWord)
			if cleanWord == "" {
				continue
			}
			if !dictionary[cleanWord] {
				if strings.HasSuffix(cleanWord, "s") && dictionary[strings.TrimSuffix(cleanWord, "s")] {
					continue
				}
				unknowns[cleanWord] = true
			}
		}

		if len(unknowns) == 0 {
			showModal("Spell Check", "No misspellings found!")
		} else {
			var list []string
			for w := range unknowns {
				list = append(list, w)
			}
			displayLimit := 20
			msg := "Potential misspellings:\n\n"
			count := 0
			for _, w := range list {
				msg += fmt.Sprintf("- %s\n", w)
				count++
				if count >= displayLimit {
					msg += fmt.Sprintf("...and %d more.", len(list)-displayLimit)
					break
				}
			}
			showModal("Spell Check Results", msg)
		}
	}

	// --- FILE IO ---
	saveBook := func(filename string, silent bool) {
		saveCurrentChapter()
		saveCurrentWiki()

		if filename == "" {
			filename = appState.CurrentFilename()
		}
		if filename == "" {
			if !silent {
				showModal("Error", "Please provide a filename: 'save <name>'")
			}
			return
		}

		proj := appState.Snapshot()
		if err := persistence.SaveProject(filename, proj); err != nil {
			if !silent {
				showModal("Error", err.Error())
			}
			return
		}
		appState.SetCurrentFilename(filename)
		if silent {
			flashStatusMessage(fmt.Sprintf(" [Autosaved to %s at %s] ", filename, time.Now().Format("15:04:05")))
		} else {
			showModal("Success", fmt.Sprintf("Saved to %s", filename))
		}
	}

	loadBook := func(filename string) {
		appState.SetLoading(true)
		defer func() { appState.SetLoading(false) }()

		proj, err := persistence.LoadProject(filename)
		if err != nil {
			showModal("Error", err.Error())
			return
		}

		appState.SetProject(proj)
		appState.SetCurrentFilename(filename)
		appState.SetView(ViewMain)
		appState.SetFocusMode(false)

		setView(ViewMain)
		loadChapter(0)

		showModal("Success", fmt.Sprintf("Loaded %s", filename))
	}

	go func() {
		ticker := time.NewTicker(60 * time.Second)
		for range ticker.C {
			if appState.Loading() {
				continue
			}
			fname := appState.CurrentFilename()
			if fname != "" {
				app.QueueUpdateDraw(func() { saveBook(fname, true) })
			}
		}
	}()

	// --- FILE PICKER ---
	showFilePicker := func() {
		// Get list of .json files in current directory
		files, err := os.ReadDir(".")
		if err != nil {
			showModal("Error", "Could not read directory")
			return
		}

		var jsonFiles []string
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
				jsonFiles = append(jsonFiles, file.Name())
			}
		}

		if len(jsonFiles) == 0 {
			showModal("No Files", "No .json files found in current directory.\nUsage: open <filename>")
			return
		}

		// Create file picker list
		fileList := tview.NewList()
		fileList.ShowSecondaryText(false)
		fileList.SetHighlightFullLine(true)
		fileList.SetSelectedBackgroundColor(tview.Styles.TitleColor)
		fileList.SetSelectedTextColor(tview.Styles.PrimitiveBackgroundColor)
		fileList.SetBorder(true)
		fileList.SetTitle("Open file (↑↓ to navigate)")
		fileList.SetBorderPadding(1, 1, 2, 2)

		// Add files to list
		for _, filename := range jsonFiles {
			fname := filename // Capture for closure
			fileList.AddItem(fname, "", 0, func() {
				pages.HidePage("filepicker")
				loadBook(fname)
			})
		}

		// Handle escape key
		fileList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEscape {
				pages.HidePage("filepicker")
				switch appState.CurrentView() {
				case ViewNotes:
					app.SetFocus(notesArea)
				case ViewWiki:
					app.SetFocus(wikiArea)
				default:
					app.SetFocus(textArea)
				}
				return nil
			}
			return event
		})

		// Show the file picker
		pages.AddPage("filepicker", tview.NewGrid().
			SetColumns(0, 60, 0).
			SetRows(0, 20, 0).
			AddItem(fileList, 1, 1, 1, 1, 0, 0, true), true, true)
		app.SetFocus(fileList)
	}

	// --- COMMAND PROCESSING ---
	handleCommand := func(cmdRaw string) {
		cmdRaw = strings.TrimSpace(cmdRaw)
		parts := strings.Fields(cmdRaw)
		if len(parts) == 0 {
			return
		}
		cmd := strings.ToLower(parts[0])

		// Delegate to shared command registry where possible
		if handler, ok := commands.Registry[cmd]; ok {
			if (cmd == "open" || cmd == "load") && len(parts) == 1 {
				showFilePicker()
				return
			}
			if cmd == "save" || cmd == "export" {
				saveCurrentChapter()
				saveCurrentWiki()
			}

			res := handler(parts[1:], appState)
			if res.Err != nil {
				showModal("Error", res.Err.Error())
				return
			}

			if cmd == "open" || cmd == "load" {
				appState.SetView(ViewMain)
				appState.SetFocusMode(false)
				setView(ViewMain)
				loadChapter(appState.CurrentChapterIndex())
			}

			if res.Modal {
				showModal("Success", res.Message)
			} else if res.Message != "" {
				flashStatusMessage(res.Message)
			}
			return
		}

		switch cmd {
		case "quit", "exit":
			app.Stop()
		case "help":
			pages.ShowPage("help")
		case "main", "edit":
			setView(ViewMain)
		case "wordcount":
			targetArea := textArea
			view := appState.CurrentView()
			if view == ViewNotes {
				targetArea = notesArea
			} else if view == ViewWiki {
				targetArea = wikiArea
			}
			text := targetArea.GetText()
			words := len(strings.Fields(text))
			lines := strings.Count(text, "\n") + 1
			if len(text) == 0 {
				lines = 0
			}
			showModal("Stats", fmt.Sprintf("Words: %d\nChars: %d\nLines: %d", words, len(text), lines))
		case "chapters", "list":
			// Explicit list creation to avoid chaining errors
			list := tview.NewList()
			list.ShowSecondaryText(false)
			list.SetHighlightFullLine(true)
			list.SetSelectedBackgroundColor(tview.Styles.TitleColor)
			list.SetSelectedTextColor(tview.Styles.PrimitiveBackgroundColor)
			list.SetBorder(true)
			list.SetTitle("Chapters (< & > reorder)")
			list.SetBorderPadding(1, 1, 2, 2)

			proj := appState.Snapshot()
			currentIdx := appState.CurrentChapterIndex()
			for i, chap := range proj.Chapters {
				idx := i
				title := fmt.Sprintf("%d. %s", i+1, chap.Title)
				if i == currentIdx {
					title += " (Current)"
				}
				list.AddItem(title, "", 0, func() { loadChapter(idx) })
			}

			list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyEscape {
					pages.HidePage("modal")
					app.SetFocus(textArea)
					return nil
				}
				return event
			})

			grid := tview.NewGrid().SetColumns(0, 40, 0).SetRows(0, 20, 0).AddItem(list, 1, 1, 1, 1, 0, 0, true)
			pages.AddPage("modal", grid, true, true)
			app.SetFocus(list)
		case "search":
			if len(parts) > 1 {
				term := strings.Join(parts[1:], " ")
				targetArea := textArea
				view := appState.CurrentView()
				if view == ViewNotes {
					targetArea = notesArea
				} else if view == ViewWiki {
					targetArea = wikiArea
				}
				count := strings.Count(targetArea.GetText(), term)
				showModal("Search", fmt.Sprintf("Found %d of '%s'", count, term))
			}
		case "spellcheck", "spell":
			runSpellCheck()
		case "theme":
			if len(parts) > 1 {
				applyTheme(parts[1])
			}
		case "notes":
			toggleNotes()
		case "analyze":
			runAnalysis()

		// Import plain text into chapter
		case "import":
			// Usage:
			//   import <file.txt>         -> overwrite current chapter with file contents
			//   import new <file.txt>     -> create a new chapter with file contents (title = filename)
			if len(parts) < 2 {
				showModal("Error", "Usage: import <file.txt>  OR  import new <file.txt>")
				break
			}

			// helper to validate .txt path
			validateTxt := func(raw string) (string, bool) {
				fn := strings.Join(strings.Fields(raw), " ")
				fn = filepath.Clean(fn)
				if !strings.HasSuffix(strings.ToLower(fn), ".txt") {
					return fn, false
				}
				return fn, true
			}

			if parts[1] == "new" {
				if len(parts) < 3 {
					showModal("Error", "Usage: import new <file.txt>")
					break
				}
				fn, ok := validateTxt(strings.Join(parts[2:], " "))
				if !ok {
					showModal("Error", "Only .txt files supported for import.")
					break
				}

				flashStatusMessage("Importing file...")

				// Read file off the UI goroutine, update UI via QueueUpdateDraw
				go func(path string) {
					data, err := os.ReadFile(path)
					app.QueueUpdateDraw(func() {
						if err != nil {
							showModal("Error", fmt.Sprintf("Failed to read file: %v", err))
							return
						}

						title := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

						newIdx := appState.NewChapter(title)
						appState.SaveCurrentChapter(string(data), "")
						loadChapter(newIdx)
						flashStatusMessage(fmt.Sprintf("Imported %s into new chapter '%s'", path, title))
					})
				}(fn)

				break
			}

			// default: import into current chapter (overwrite)
			fn, ok := validateTxt(strings.Join(parts[1:], " "))
			if !ok {
				showModal("Error", "Only .txt files supported for import.")
				break
			}

			flashStatusMessage("Importing file...")

			go func(path string) {
				data, err := os.ReadFile(path)
				app.QueueUpdateDraw(func() {
					if err != nil {
						showModal("Error", fmt.Sprintf("Failed to read file: %v", err))
						return
					}

					content := string(data)
					proj := appState.Snapshot()
					currentIdx := appState.CurrentChapterIndex()

					if currentIdx < 0 || currentIdx >= len(proj.Chapters) {
						title := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
						newIdx := appState.NewChapter(title)
						appState.SaveCurrentChapter(content, "")
						loadChapter(newIdx)
						flashStatusMessage(fmt.Sprintf("Imported %s into new chapter '%s'", path, title))
						return
					}

					notes := notesArea.GetText()
					appState.SaveCurrentChapter(content, notes)
					updated := appState.Snapshot()
					title := updated.Chapters[currentIdx].Title
					textArea.SetText(content, false)
					textArea.SetTitle(fmt.Sprintf("gowrite - Chapter %d: %s", currentIdx+1, title))

					flashStatusMessage(fmt.Sprintf("Imported %s into Chapter %d", path, currentIdx+1))
				})
			}(fn)

			// WIKI COMMANDS
		case "wiki":
			if len(parts) > 1 {
				sub := strings.ToLower(parts[1])

				if sub == "new" {
					title := "New Entry"
					if len(parts) > 2 {
						title = strings.Join(parts[2:], " ")
					}
					newIdx := appState.NewWiki(title)
					loadWiki(newIdx)
					setView(ViewWiki)
				} else if sub == "delete" {
					deleteWiki(appState.CurrentWikiIndex())
				} else if sub == "rename" {
					if len(parts) > 2 {
						renameWiki(appState.CurrentWikiIndex(), strings.Join(parts[2:], " "))
					}
				} else {
					// Assume they typed 'wiki searchterm' or similar, but for now just open view
					setView(ViewWiki)
				}
			} else {
				setView(ViewWiki)
			}

		case "structure":
			if len(parts) > 1 {
				applyStructure(parts[1])
			} else {
				showModal("Structure", "Usage: structure <name>\nOptions: 3act, hero, cat, fichtean, horror")
			}

		case "chapter":
			if len(parts) > 1 {
				sub := strings.ToLower(parts[1])
				if sub == "new" {
					title := "New Chapter"
					if len(parts) > 2 {
						title = strings.Join(parts[2:], " ")
					}
					saveCurrentChapter()
					newIdx := appState.NewChapter(title)
					loadChapter(newIdx)
				} else if sub == "delete" {
					idx := appState.CurrentChapterIndex()
					if len(parts) > 2 {
						if n, err := strconv.Atoi(parts[2]); err == nil {
							idx = n - 1
						}
					}
					deleteChapter(idx)
				} else if sub == "rename" {
					// Supports 'chapter rename Title' (current) or 'chapter rename 1 Title'
					idx := appState.CurrentChapterIndex()
					nameStart := 2
					if len(parts) > 2 {
						// Check if first arg is a number
						if n, err := strconv.Atoi(parts[2]); err == nil {
							idx = n - 1
							nameStart = 3
						}
					}
					if nameStart < len(parts) {
						renameChapter(idx, strings.Join(parts[nameStart:], " "))
					}
				}
			}
		}
	}

	updateInfos := func() {
		view := appState.CurrentView()
		if view == ViewAnalyze {
			position.SetText(" Read-Only ")
			return
		}

		targetArea := textArea
		if view == ViewNotes {
			targetArea = notesArea
		} else if view == ViewWiki {
			targetArea = wikiArea
		}

		fromRow, fromColumn, _, _ := targetArea.GetCursor()
		text := targetArea.GetText()
		wordCount := len(strings.Fields(text))

		wordCountStr := fmt.Sprintf("[%s]%d[white]", tview.Styles.SecondaryTextColor, wordCount)
		position.SetText(fmt.Sprintf("Words: %s | Row: %d Col: %d ", wordCountStr, fromRow, fromColumn))
	}
	textArea.SetMovedFunc(updateInfos)
	notesArea.SetMovedFunc(updateInfos)
	wikiArea.SetMovedFunc(updateInfos)
	updateInfos()

	commandPalette.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			cmd := commandPalette.GetText()
			commandPalette.SetText("")
			handleCommand(cmd)

			// Intelligent focus restoration
			isModal := false
			for _, m := range []string{"help", "chapters", "list", "wordcount", "save", "open", "load", "export", "search", "replace", "spell", "theme", "analyze", "target", "chapter", "wiki", "structure", "import"} {
				if strings.HasPrefix(cmd, m) {
					isModal = true
					break
				}
			}
			if !isModal {
				switch appState.CurrentView() {
				case ViewNotes:
					app.SetFocus(notesArea)
				case ViewAnalyze:
					app.SetFocus(analysisView)
				case ViewWiki:
					app.SetFocus(wikiArea)
				default:
					app.SetFocus(textArea)
				}
			}
		} else if key == tcell.KeyEscape {
			commandPalette.SetText("")
			switch appState.CurrentView() {
			case ViewNotes:
				app.SetFocus(notesArea)
			case ViewAnalyze:
				app.SetFocus(analysisView)
			case ViewWiki:
				app.SetFocus(wikiArea)
			default:
				app.SetFocus(textArea)
			}
		}
	})

	// WIKI INPUT CAPTURE
	// We do NOT capture Enter here, allowing the List to handle selection logic naturally
	wikiList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			app.SetFocus(wikiArea)
			return nil
		}
		// Removed Esc handler here because global Esc or Ctrl-W handles it better
		return event
	})

	wikiArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			saveCurrentWiki()
			app.SetFocus(wikiList)
			return nil
		}
		// Removed Esc handler here too
		return event
	})

	// --- 5. HELP ---
	help1 := tview.NewTextView()
	help1.SetDynamicColors(true)
	help1.SetText(`[green]Navigation
[yellow]Arrows[white]: Move cursor
[yellow]Ctrl-A/Home[white]: Start of line
[yellow]Ctrl-E/End[white]: End of line
[blue]Enter for next page, Esc to return.`)

	help2 := tview.NewTextView()
	help2.SetDynamicColors(true)
	help2.SetText(`[green]Editing & View
Type to enter text.
[yellow]Ctrl-Q[white]: Copy | [yellow]Ctrl-X[white]: Cut | [yellow]Ctrl-V[white]: Paste
[yellow]Ctrl-Z[white]: Undo | [yellow]Ctrl-Y[white]: Redo
[yellow]Ctrl-T[white]: Toggle Center View
[yellow]Ctrl-F[white]: Toggle Focus Mode
[blue]Enter for next page, Esc to return.`)

	helpCmds := tview.NewTextView()
	helpCmds.SetDynamicColors(true)
	helpCmds.SetText(`[green]Commands (Ctrl-E)
[yellow]structure <type>[white]: Apply template (3act, hero, cat, fichtean, horror)
[yellow]wiki[white]: Open Story Wiki (Ctrl-W to close)
[yellow]wiki new <name>[white]: Add entry
[yellow]wiki rename <name>[white]: Rename entry
[yellow]wiki delete[white]: Delete entry
[yellow]save <file>[white]: Save project
[yellow]open[white]: Show file picker (or [yellow]open <file>[white] to open directly)
[yellow]export <file>[white]: Export to text
[yellow]notes[white] (or Ctrl-N): Toggle Notes
[yellow]analyze[white]: Hemingway Analysis Mode
[yellow]chapter new/delete/rename[white]: Manage chapters
[yellow]import <file.txt>[white]: Import .txt into current chapter
[yellow]import new <file.txt>[white]: Import .txt into a new chapter`)

	// Setup the frame for Help pages
	help := tview.NewFrame(help1)
	help.SetBorders(1, 1, 0, 0, 2, 2)
	help.SetTitle("Help")

	// State tracking for help pagination
	helpPageIndex := 0

	help.SetBorder(true)
	help.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		if e.Key() == tcell.KeyEscape {
			pages.SwitchToPage("main")
			// Reset help state
			help.SetPrimitive(help1)
			helpPageIndex = 0

			// Restore focus
			switch appState.CurrentView() {
			case ViewNotes:
				app.SetFocus(notesArea)
			case ViewAnalyze:
				app.SetFocus(analysisView)
			case ViewWiki:
				app.SetFocus(wikiArea)
			default:
				app.SetFocus(textArea)
			}
			return nil
		}
		if e.Key() == tcell.KeyEnter {
			// Cycle through pages
			helpPageIndex = (helpPageIndex + 1) % 3
			switch helpPageIndex {
			case 0:
				help.SetPrimitive(help1)
			case 1:
				help.SetPrimitive(help2)
			case 2:
				help.SetPrimitive(helpCmds)
			}
			return nil
		}
		return e
	})

	pages.AddAndSwitchToPage("main", mainView, true)
	pages.AddPage("help", tview.NewGrid().SetColumns(0, 64, 0).SetRows(0, 22, 0).AddItem(help, 1, 1, 1, 1, 0, 0, true), true, false)

	// --- ANALYSIS INPUT CAPTURE ---
	analysisView.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		if e.Key() == tcell.KeyEscape {
			setView(ViewMain) // Return to editor on Esc
			return nil
		}
		return e
	})

	// --- GLOBAL KEYS ---
	app.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		if e.Key() == tcell.KeyF1 {
			pages.ShowPage("help")
			return nil
		}
		if e.Key() == tcell.KeyCtrlT {
			appState.SetCentered(!appState.Centered())
			app.ForceDraw()
			return nil
		}
		// FOCUS MODE TOGGLE (Ctrl-F)
		if e.Key() == tcell.KeyCtrlF {
			toggleFocus()
			return nil
		}
		// WIKI TOGGLE (Ctrl-W)
		if e.Key() == tcell.KeyCtrlW {
			toggleWiki()
			return nil
		}
		if e.Key() == tcell.KeyCtrlG {
			handleCommand("chapters")
			return nil
		}
		if e.Key() == tcell.KeyCtrlE {
			// Auto-exit Focus Mode if user wants to run a command
			if appState.FocusMode() {
				toggleFocus()
			}
			if app.GetFocus() != commandPalette {
				app.SetFocus(commandPalette)
			} else {
				switch appState.CurrentView() {
				case ViewNotes:
					app.SetFocus(notesArea)
				case ViewAnalyze:
					app.SetFocus(analysisView)
				case ViewWiki:
					app.SetFocus(wikiArea)
				default:
					app.SetFocus(textArea)
				}
			}
			return nil
		}
		if e.Key() == tcell.KeyCtrlS {
			saveBook(appState.CurrentFilename(), false)
			return nil
		}
		// Ctrl-N Handler
		if e.Key() == tcell.KeyCtrlN {
			toggleNotes()
			return nil
		}
		return e
	})

	if err := app.SetRoot(pages, true).EnableMouse(true).EnablePaste(true).Run(); err != nil {
		panic(err)
	}
}
