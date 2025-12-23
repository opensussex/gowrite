// gowrite - A distraction-free writing tool with Hemingway Analysis and Story Bible
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Chapter represents a section of the document
type Chapter struct {
	Title   string
	Content string
	Notes   string
	Target  int
}

// WikiEntry represents a single item in the Story Bible
type WikiEntry struct {
	Title   string
	Content string
}

// Project represents the full save file structure (Chapters + Wiki)
type Project struct {
	Chapters []Chapter
	Wiki     []WikiEntry
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

	chapters := []Chapter{
		{Title: "The Beginning", Content: "", Notes: "", Target: 0},
	}

	wikiEntries := []WikiEntry{
		{Title: "General Notes", Content: ""},
	}

	currentChapterIndex := 0
	currentWikiIndex := 0
	currentFilename := ""

	// View States
	const (
		ViewMain = iota
		ViewNotes
		ViewAnalyze
		ViewWiki
	)
	currentView := ViewMain

	// Visual States
	isCenteredView := false
	isFocusMode := false // Hides all UI chrome
	const TargetWidth = 85

	dictionary := make(map[string]bool)
	dictionaryLoaded := false

	// --- 2. Setup Main Components ---

	// MAIN EDITOR
	textArea := tview.NewTextArea()
	textArea.SetWrap(true)
	textArea.SetPlaceholder("Start writing your masterpiece...")
	textArea.SetTextStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite))
	textArea.SetTitle(fmt.Sprintf("gowrite - Chapter 1: %s", chapters[0].Title))
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

	// WIKI LIST (Story Bible)
	wikiList := tview.NewList()
	wikiList.ShowSecondaryText(false)
	wikiList.SetBorder(true)
	wikiList.SetTitle("Story Bible (Ctrl-W to Close)")
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

	defaultHelpText := " F1: Help | Ctrl-N: Notes | Ctrl-W: Wiki | Ctrl-T: Center | Ctrl-E: Command"
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
	applyTheme("dark")

	// --- 4. Logic & Helper Functions ---

	// VIEW RESIZE LOGIC
	app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		w, _ := screen.Size()

		var hPadding int
		// If centered view is ON and screen is wide enough to justify it
		if isCenteredView && w > TargetWidth+4 {
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

	saveCurrentChapter := func() {
		if currentChapterIndex >= 0 && currentChapterIndex < len(chapters) {
			chapters[currentChapterIndex].Content = textArea.GetText()
			chapters[currentChapterIndex].Notes = notesArea.GetText()
		}
	}

	saveCurrentWiki := func() {
		if len(wikiEntries) > 0 && currentWikiIndex < len(wikiEntries) {
			wikiEntries[currentWikiIndex].Content = wikiArea.GetText()
		}
	}

	loadChapter := func(index int) {
		saveCurrentChapter()
		currentChapterIndex = index
		chapter := chapters[index]

		textArea.SetText(chapter.Content, false)
		notesArea.SetText(chapter.Notes, false)

		title := fmt.Sprintf("gowrite - Chapter %d: %s", index+1, chapter.Title)
		if currentView == ViewNotes {
			title += " (NOTES)"
		}
		textArea.SetTitle(title)
		notesArea.SetTitle(fmt.Sprintf("NOTES - Chapter %d", index+1))

		pages.HidePage("modal")

		if currentView == ViewNotes {
			app.SetFocus(notesArea)
		} else {
			app.SetFocus(textArea)
		}
	}

	// Forward declaration for recursion
	var loadWiki func(int)

	loadWiki = func(index int) {
		saveCurrentWiki()
		if index < 0 || index >= len(wikiEntries) {
			return
		}
		currentWikiIndex = index
		entry := wikiEntries[index]

		wikiArea.SetText(entry.Content, false)
		wikiArea.SetTitle(fmt.Sprintf("Wiki: %s", entry.Title))

		// Populate List
		wikiList.Clear()
		for i, w := range wikiEntries {
			title := w.Title
			if i == currentWikiIndex {
				title += " *"
			}
			idx := i
			// Note: The callback here handles both Loading AND Focusing
			wikiList.AddItem(title, "", 0, func() {
				loadWiki(idx)
				app.SetFocus(wikiArea)
			})
		}
		wikiList.SetCurrentItem(currentWikiIndex)
	}

	setView := func(viewType int) {
		// Save state before switching
		if currentView == ViewWiki {
			saveCurrentWiki()
		} else {
			saveCurrentChapter()
		}

		currentView = viewType

		// 1. Clear grid completely
		mainView.Clear()

		// 2. Determine which widget to show
		var activeWidget tview.Primitive
		var title string
		chapter := chapters[currentChapterIndex]

		switch viewType {
		case ViewMain:
			activeWidget = textArea
			title = fmt.Sprintf("gowrite - Chapter %d: %s", currentChapterIndex+1, chapter.Title)
			helpInfo.SetText(defaultHelpText)
			mainView.SetColumns(0) // Reset to single column

		case ViewNotes:
			activeWidget = notesArea
			title = fmt.Sprintf("gowrite - Chapter %d: %s (NOTES)", currentChapterIndex+1, chapter.Title)
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
			title = "Story Bible"
			helpInfo.SetText(" WIKI | Enter: Select | Tab: Edit Text | Ctrl-W: Close | 'wiki new/del' to manage")

			loadWiki(currentWikiIndex)

			// Setup split screen
			mainView.SetColumns(30, 0)
			mainView.SetRows(0, 3, 1)

			mainView.AddItem(wikiList, 0, 0, 1, 1, 0, 0, true)
			mainView.AddItem(wikiArea, 0, 1, 1, 1, 0, 0, false)
			mainView.AddItem(commandPalette, 1, 0, 1, 2, 0, 0, false)
			mainView.AddItem(helpInfo, 2, 0, 1, 1, 0, 0, false)
			mainView.AddItem(position, 2, 1, 1, 1, 0, 0, false)

			// Wiki-specific Focus Mode
			if isFocusMode {
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
		if isFocusMode {
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
		if currentView == ViewNotes || currentView == ViewWiki {
			setView(ViewMain)
		} else {
			setView(ViewNotes)
		}
	}

	toggleWiki := func() {
		if currentView == ViewWiki {
			setView(ViewMain)
		} else {
			setView(ViewWiki)
		}
	}

	toggleFocus := func() {
		isFocusMode = !isFocusMode
		setView(currentView) // Refreshes layout
	}

	showModal := func(title, text string) {
		modal := tview.NewModal()
		modal.SetText(text)
		modal.AddButtons([]string{"OK"})
		modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.HidePage("modal")
			// Restore focus
			if currentView == ViewNotes {
				app.SetFocus(notesArea)
			} else if currentView == ViewAnalyze {
				app.SetFocus(analysisView)
			} else if currentView == ViewWiki {
				app.SetFocus(wikiArea)
			} else {
				app.SetFocus(textArea)
			}
		})

		modal.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
			if e.Key() == tcell.KeyEnter {
				pages.HidePage("modal")
				if currentView == ViewNotes {
					app.SetFocus(notesArea)
				} else if currentView == ViewWiki {
					app.SetFocus(wikiArea)
				} else {
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

	showYesNoModal := func(title, text string, onYes func()) {
		modal := tview.NewModal()
		modal.SetText(text)
		modal.AddButtons([]string{"Yes", "No"})
		modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				onYes()
			}
			pages.HidePage("modal")
			if currentView == ViewNotes {
				app.SetFocus(notesArea)
			} else if currentView == ViewWiki {
				app.SetFocus(wikiArea)
			} else {
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
		if len(chapters) <= 1 {
			showModal("Error", "Cannot delete only chapter.")
			return
		}
		if index < 0 || index >= len(chapters) {
			showModal("Error", "Invalid chapter.")
			return
		}

		showYesNoModal("Confirm", fmt.Sprintf("Delete Chapter %d?", index+1), func() {
			chapters = append(chapters[:index], chapters[index+1:]...)
			if index < currentChapterIndex {
				currentChapterIndex--
			} else if index == currentChapterIndex && currentChapterIndex >= len(chapters) {
				currentChapterIndex = len(chapters) - 1
			}
			loadChapter(currentChapterIndex)
		})
	}

	renameChapter := func(index int, newName string) {
		if index < 0 || index >= len(chapters) {
			return
		}
		chapters[index].Title = newName
		if index == currentChapterIndex {
			loadChapter(currentChapterIndex)
		} else {
			showModal("Success", fmt.Sprintf("Renamed Chapter %d to '%s'", index+1, newName))
		}
	}

	// --- WIKI OPS ---
	deleteWiki := func(index int) {
		if len(wikiEntries) <= 1 {
			showModal("Error", "Cannot delete the only wiki entry.")
			return
		}
		showYesNoModal("Confirm", fmt.Sprintf("Delete Wiki Entry '%s'?", wikiEntries[index].Title), func() {
			wikiEntries = append(wikiEntries[:index], wikiEntries[index+1:]...)
			if index < currentWikiIndex {
				currentWikiIndex--
			} else if index == currentWikiIndex && currentWikiIndex >= len(wikiEntries) {
				currentWikiIndex = len(wikiEntries) - 1
			}
			loadWiki(currentWikiIndex)
		})
	}

	renameWiki := func(index int, newName string) {
		if index < 0 || index >= len(wikiEntries) {
			return
		}
		wikiEntries[index].Title = newName
		loadWiki(currentWikiIndex)
	}

	// --- ANALYSIS LOGIC (Hemingway) ---

	calculateReadability := func(text string) string {
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

		return fmt.Sprintf("Grade %d", grade)
	}

	runAnalysis := func() {
		text := textArea.GetText()

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

		analysisView.SetText(processedText.String())
		setView(ViewAnalyze)

		stats := calculateReadability(text)
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
		if currentView == ViewNotes {
			targetArea = notesArea
		} else if currentView == ViewWiki {
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
		saveCurrentWiki() // Save Wiki entries too

		if filename == "" {
			if currentFilename == "" {
				if !silent {
					showModal("Error", "Please provide a filename: 'save <name>'")
				}
				return
			}
			filename = currentFilename
		}
		if !strings.HasSuffix(filename, ".json") {
			filename += ".json"
		}

		// Create project struct to hold both Chapters and Wiki
		projectData := Project{
			Chapters: chapters,
			Wiki:     wikiEntries,
		}

		data, err := json.MarshalIndent(projectData, "", "  ")
		if err != nil {
			if !silent {
				showModal("Error", err.Error())
			}
			return
		}

		err = os.WriteFile(filename, data, 0644)
		if err != nil {
			if !silent {
				showModal("Error", err.Error())
			}
			return
		}

		currentFilename = filename
		if silent {
			flashStatusMessage(fmt.Sprintf(" [Autosaved to %s at %s] ", filename, time.Now().Format("15:04:05")))
		} else {
			showModal("Success", fmt.Sprintf("Saved to %s", filename))
		}
	}

	loadBook := func(filename string) {
		if !strings.HasSuffix(filename, ".json") {
			filename += ".json"
		}
		data, err := os.ReadFile(filename)
		if err != nil {
			showModal("Error", err.Error())
			return
		}

		// Try loading as Project struct (New format)
		var projectData Project
		err = json.Unmarshal(data, &projectData)

		validLoad := false

		if err == nil && len(projectData.Chapters) > 0 {
			// Success: It's the new format
			chapters = projectData.Chapters
			wikiEntries = projectData.Wiki
			validLoad = true
		} else {
			// Failure: Try loading as old format (Just Array of Chapters)
			var oldChapters []Chapter
			if err2 := json.Unmarshal(data, &oldChapters); err2 == nil && len(oldChapters) > 0 {
				chapters = oldChapters
				wikiEntries = []WikiEntry{{Title: "General", Content: ""}} // Default wiki
				validLoad = true
			}
		}

		if !validLoad {
			showModal("Error", "File empty or corrupt.")
			return
		}

		// Ensure Wiki isn't empty if loading from old file
		if len(wikiEntries) == 0 {
			wikiEntries = []WikiEntry{{Title: "General", Content: ""}}
		}

		// STATE RESET
		currentFilename = filename
		currentChapterIndex = 0
		currentWikiIndex = 0
		currentView = ViewMain
		isFocusMode = false

		setView(ViewMain)
		loadChapter(0)

		showModal("Success", fmt.Sprintf("Loaded %s", filename))
	}

	exportBook := func(filename string) {
		saveCurrentChapter()
		if filename == "" {
			showModal("Error", "Usage: export <filename>")
			return
		}
		if !strings.Contains(filename, ".") {
			filename += ".txt"
		}

		var sb strings.Builder
		for i, chap := range chapters {
			sb.WriteString(fmt.Sprintf("# Chapter %d: %s\n\n", i+1, chap.Title))
			sb.WriteString(chap.Content)
			sb.WriteString("\n\n")
		}
		if err := os.WriteFile(filename, []byte(sb.String()), 0644); err != nil {
			showModal("Error", err.Error())
		} else {
			showModal("Success", fmt.Sprintf("Exported to %s", filename))
		}
	}

	go func() {
		ticker := time.NewTicker(60 * time.Second)
		for range ticker.C {
			if currentFilename != "" {
				app.QueueUpdateDraw(func() { saveBook(currentFilename, true) })
			}
		}
	}()

	// --- COMMAND PROCESSING ---
	handleCommand := func(cmdRaw string) {
		cmdRaw = strings.TrimSpace(cmdRaw)
		parts := strings.Fields(cmdRaw)
		if len(parts) == 0 {
			return
		}
		cmd := strings.ToLower(parts[0])

		switch cmd {
		case "quit", "exit":
			app.Stop()
		case "help":
			pages.ShowPage("help")
		case "main", "edit":
			setView(ViewMain)
		case "wordcount":
			targetArea := textArea
			if currentView == ViewNotes {
				targetArea = notesArea
			} else if currentView == ViewWiki {
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

			// Simple populate logic
			for i, chap := range chapters {
				idx := i
				title := fmt.Sprintf("%d. %s", i+1, chap.Title)
				if i == currentChapterIndex {
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

		case "save":
			f := ""
			if len(parts) > 1 {
				f = strings.Join(parts[1:], " ")
			}
			saveBook(f, false)
		case "open", "load":
			if len(parts) > 1 {
				loadBook(strings.Join(parts[1:], " "))
			} else {
				showModal("Error", "Usage: open <file>")
			}
		case "export":
			if len(parts) > 1 {
				exportBook(strings.Join(parts[1:], " "))
			} else {
				showModal("Error", "Usage: export <file>")
			}
		case "search":
			if len(parts) > 1 {
				term := strings.Join(parts[1:], " ")
				targetArea := textArea
				if currentView == ViewNotes {
					targetArea = notesArea
				} else if currentView == ViewWiki {
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

		// WIKI COMMANDS
		case "wiki":
			if len(parts) > 1 {
				sub := strings.ToLower(parts[1])

				if sub == "new" {
					title := "New Entry"
					if len(parts) > 2 {
						title = strings.Join(parts[2:], " ")
					}
					wikiEntries = append(wikiEntries, WikiEntry{Title: title, Content: ""})
					currentWikiIndex = len(wikiEntries) - 1
					setView(ViewWiki)
				} else if sub == "delete" {
					deleteWiki(currentWikiIndex)
				} else if sub == "rename" {
					if len(parts) > 2 {
						renameWiki(currentWikiIndex, strings.Join(parts[2:], " "))
					}
				} else {
					// Assume they typed 'wiki searchterm' or similar, but for now just open view
					setView(ViewWiki)
				}
			} else {
				setView(ViewWiki)
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
					chapters = append(chapters, Chapter{Title: title})
					loadChapter(len(chapters) - 1)
				} else if sub == "delete" {
					idx := currentChapterIndex
					if len(parts) > 2 {
						if n, err := strconv.Atoi(parts[2]); err == nil {
							idx = n - 1
						}
					}
					deleteChapter(idx)
				} else if sub == "rename" {
					// Supports 'chapter rename Title' (current) or 'chapter rename 1 Title'
					idx := currentChapterIndex
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
		if currentView == ViewAnalyze {
			position.SetText(" Read-Only ")
			return
		}

		targetArea := textArea
		if currentView == ViewNotes {
			targetArea = notesArea
		} else if currentView == ViewWiki {
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
			for _, m := range []string{"help", "chapters", "list", "wordcount", "save", "open", "load", "export", "search", "replace", "spell", "theme", "analyze", "target", "chapter", "wiki"} {
				if strings.HasPrefix(cmd, m) {
					isModal = true
					break
				}
			}
			if !isModal {
				if currentView == ViewNotes {
					app.SetFocus(notesArea)
				} else if currentView == ViewAnalyze {
					app.SetFocus(analysisView)
				} else if currentView == ViewWiki {
					app.SetFocus(wikiArea)
				} else {
					app.SetFocus(textArea)
				}
			}
		} else if key == tcell.KeyEscape {
			commandPalette.SetText("")
			if currentView == ViewNotes {
				app.SetFocus(notesArea)
			} else if currentView == ViewAnalyze {
				app.SetFocus(analysisView)
			} else if currentView == ViewWiki {
				app.SetFocus(wikiArea)
			} else {
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
[yellow]wiki[white]: Open Story Bible (Ctrl-W to close)
[yellow]wiki new <name>[white]: Add entry
[yellow]wiki rename <name>[white]: Rename entry
[yellow]wiki delete[white]: Delete entry
[yellow]save/open/export <file>[white]: File ops
[yellow]notes[white] (or Ctrl-N): Toggle Notes
[yellow]analyze[white]: Hemingway Analysis Mode
[yellow]chapter new/delete/rename[white]: Manage chapters`)

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
			if currentView == ViewNotes {
				app.SetFocus(notesArea)
			} else if currentView == ViewAnalyze {
				app.SetFocus(analysisView)
			} else if currentView == ViewWiki {
				app.SetFocus(wikiArea)
			} else {
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
			isCenteredView = !isCenteredView
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
		if e.Key() == tcell.KeyCtrlE {
			// Auto-exit Focus Mode if user wants to run a command
			if isFocusMode {
				toggleFocus()
			}
			if app.GetFocus() != commandPalette {
				app.SetFocus(commandPalette)
			} else {
				if currentView == ViewNotes {
					app.SetFocus(notesArea)
				} else if currentView == ViewAnalyze {
					app.SetFocus(analysisView)
				} else if currentView == ViewWiki {
					app.SetFocus(wikiArea)
				} else {
					app.SetFocus(textArea)
				}
			}
			return nil
		}
		if e.Key() == tcell.KeyCtrlS {
			saveBook(currentFilename, false)
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
