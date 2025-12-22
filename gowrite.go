// gowrite - A distraction-free writing tool with Hemingway Analysis
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
	currentChapterIndex := 0
	currentFilename := ""

	// View States
	const (
		ViewMain = iota
		ViewNotes
		ViewAnalyze
	)
	currentView := ViewMain

	// Centered View State
	isCenteredView := false
	const TargetWidth = 85 // The ideal reading width in characters

	dictionary := make(map[string]bool)
	dictionaryLoaded := false

	// --- 2. Setup Main Components ---

	// MAIN EDITOR
	textArea := tview.NewTextArea().
		SetWrap(true).
		SetPlaceholder("Start writing your masterpiece...")

	textArea.SetTextStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite))
	textArea.SetTitle(fmt.Sprintf("gowrite - Chapter 1: %s", chapters[0].Title)).SetBorder(true)
	textArea.SetBorderPadding(1, 1, 2, 2)

	// NOTES EDITOR
	notesArea := tview.NewTextArea().
		SetWrap(true).
		SetPlaceholder("Scene ideas, plot points, and reminders...")

	notesArea.SetTextStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorYellow))
	notesArea.SetTitle("SCENE NOTES").SetBorder(true)
	notesArea.SetBorderPadding(1, 1, 2, 2)

	// ANALYSIS VIEWER (Read Only)
	analysisView := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true).
		SetWordWrap(true)
	analysisView.SetTitle("HEMINGWAY ANALYSIS MODE").SetBorder(true)
	analysisView.SetBorderPadding(1, 1, 2, 2)

	commandPalette := tview.NewInputField().
		SetLabel(" > ").
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldTextColor(tcell.ColorWhite).
		SetLabelColor(tcell.ColorYellow).
		SetPlaceholder("Type 'help' for commands")

	commandPalette.SetBorder(true).SetBorderPadding(0, 0, 1, 1).SetTitle("Command Palette")

	defaultHelpText := " F1: Help | Ctrl-N: Notes | Ctrl-T: Center | Ctrl-S: Save | Ctrl-E: Command"
	helpInfo := tview.NewTextView().
		SetText(defaultHelpText).
		SetTextColor(tcell.ColorDarkGray)

	position := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignRight)

	pages := tview.NewPages()

	// Layout Grid
	// Row 0 is dynamic (Main Text vs Notes vs Analysis)
	mainView := tview.NewGrid().
		SetRows(0, 3, 1).
		AddItem(textArea, 0, 0, 1, 2, 0, 0, true).
		AddItem(commandPalette, 1, 0, 1, 2, 0, 0, false).
		AddItem(helpInfo, 2, 0, 1, 1, 0, 0, false).
		AddItem(position, 2, 1, 1, 1, 0, 0, false)

	// --- 3. THEME LOGIC ---

	applyTheme := func(name string) {
		name = strings.ToLower(name)
		// Reset basics
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
			analysisView.SetBackgroundColor(tcell.ColorBlack)

			commandPalette.SetFieldBackgroundColor(tcell.ColorBlack).SetFieldTextColor(tcell.ColorWhite).SetBackgroundColor(tcell.ColorBlack)
			helpInfo.SetTextColor(tcell.ColorDarkGray).SetBackgroundColor(tcell.ColorBlack)
			position.SetBackgroundColor(tcell.ColorBlack)
		}
	}
	applyTheme("dark")

	// --- 4. Logic & Helper Functions ---

	// VIEW RESIZE LOGIC (For Center Column)
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

		return false
	})

	saveCurrentChapter := func() {
		if currentChapterIndex >= 0 && currentChapterIndex < len(chapters) {
			chapters[currentChapterIndex].Content = textArea.GetText()
			chapters[currentChapterIndex].Notes = notesArea.GetText()
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

	setView := func(viewType int) {
		saveCurrentChapter()
		currentView = viewType

		// Clear grid
		mainView.RemoveItem(textArea)
		mainView.RemoveItem(notesArea)
		mainView.RemoveItem(analysisView)

		switch viewType {
		case ViewMain:
			mainView.AddItem(textArea, 0, 0, 1, 2, 0, 0, true)
			app.SetFocus(textArea)
			helpInfo.SetText(defaultHelpText)
			chapter := chapters[currentChapterIndex]
			textArea.SetTitle(fmt.Sprintf("gowrite - Chapter %d: %s", currentChapterIndex+1, chapter.Title))

		case ViewNotes:
			mainView.AddItem(notesArea, 0, 0, 1, 2, 0, 0, true)
			app.SetFocus(notesArea)
			helpInfo.SetText(" EDITING NOTES | Ctrl-N: Back | Ctrl-T: Center | Ctrl-E: Command")
			chapter := chapters[currentChapterIndex]
			textArea.SetTitle(fmt.Sprintf("gowrite - Chapter %d: %s (NOTES)", currentChapterIndex+1, chapter.Title))

		case ViewAnalyze:
			mainView.AddItem(analysisView, 0, 0, 1, 2, 0, 0, true)
			app.SetFocus(analysisView)
			helpInfo.SetText(" ANALYSIS | [Blue]Adverbs [Green]Passive [Yellow]Hard [Red]Very Hard | Esc: Exit")
		}
	}

	toggleNotes := func() {
		if currentView == ViewNotes {
			setView(ViewMain)
		} else {
			setView(ViewNotes)
		}
	}

	showModal := func(title, text string) {
		modal := tview.NewModal().
			SetText(text).
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				pages.HidePage("modal")
				// Restore focus
				if currentView == ViewNotes {
					app.SetFocus(notesArea)
				} else if currentView == ViewAnalyze {
					app.SetFocus(analysisView)
				} else {
					app.SetFocus(textArea)
				}
			})

		modal.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
			if e.Key() == tcell.KeyEnter {
				pages.HidePage("modal")
				if currentView == ViewNotes {
					app.SetFocus(notesArea)
				} else if currentView == ViewAnalyze {
					app.SetFocus(analysisView)
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
		modal := tview.NewModal().
			SetText(text).
			AddButtons([]string{"Yes", "No"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "Yes" {
					onYes()
				}
				pages.HidePage("modal")
				if currentView == ViewNotes {
					app.SetFocus(notesArea)
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

		// Automated Readability Index (ARI)
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

	runAnalysis := func() {
		text := textArea.GetText()

		// Regex patterns
		adverbRegex := regexp.MustCompile(`(?i)\b(\w+ly)\b`)
		passiveRegex := regexp.MustCompile(`(?i)\b(am|are|is|was|were|be|been|being)\b\s+(\w+ed)\b`)

		paragraphs := strings.Split(text, "\n")
		var processedText strings.Builder

		for _, para := range paragraphs {
			if strings.TrimSpace(para) == "" {
				processedText.WriteString("\n")
				continue
			}

			// Capture sentences including delimiters
			sentenceRe := regexp.MustCompile(`[^.!?]+[.!?]*`)
			matches := sentenceRe.FindAllString(para, -1)

			for _, s := range matches {
				wordCount := len(strings.Fields(s))
				coloredS := s

				// Sentence Complexity Color
				prefix := ""
				suffix := ""

				if wordCount > 20 {
					prefix = "[red]"
					suffix = "[-]"
				} else if wordCount > 14 {
					prefix = "[yellow]"
					suffix = "[-]"
				}

				// Highlight Adverbs
				coloredS = adverbRegex.ReplaceAllStringFunc(coloredS, func(m string) string {
					return "[blue]" + m + "[-]" + prefix
				})

				// Highlight Passive
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

		data, err := json.MarshalIndent(chapters, "", "  ")
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

	loadBook := func(filename string) {
		if !strings.HasSuffix(filename, ".json") {
			filename += ".json"
		}
		data, err := os.ReadFile(filename)
		if err != nil {
			showModal("Error", err.Error())
			return
		}

		var newChapters []Chapter
		if err := json.Unmarshal(data, &newChapters); err != nil {
			showModal("Error", "Corrupt file format.")
			return
		}

		if len(newChapters) == 0 {
			showModal("Error", "File empty.")
			return
		}

		// STATE RESET - BYPASSING SAVE HANDLERS
		chapters = newChapters
		currentFilename = filename
		currentChapterIndex = 0
		currentView = ViewMain

		// Manually Update Interface (No Save Trigger)
		mainView.RemoveItem(textArea)
		mainView.RemoveItem(notesArea)
		mainView.RemoveItem(analysisView)
		mainView.AddItem(textArea, 0, 0, 1, 2, 0, 0, true)

		c := chapters[0]
		textArea.SetText(c.Content, false)
		notesArea.SetText(c.Notes, false)
		textArea.SetTitle(fmt.Sprintf("gowrite - Chapter 1: %s", c.Title))
		notesArea.SetTitle("NOTES - Chapter 1")

		app.SetFocus(textArea)
		helpInfo.SetText(defaultHelpText)

		showModal("Success", fmt.Sprintf("Loaded %s", filename))
	}

	go func() {
		ticker := time.NewTicker(60 * time.Second)
		for range ticker.C {
			if currentFilename != "" {
				app.QueueUpdateDraw(func() { saveBook(currentFilename, true) })
			}
		}
	}()

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
		loadChapter(currentChapterIndex)
	}

	setChapterTarget := func(index int, target int) {
		if index >= 0 && index < len(chapters) {
			chapters[index].Target = target
			msg := "Target removed."
			if target > 0 {
				msg = fmt.Sprintf("Target: %d", target)
			}
			showModal("Target", msg)
		}
	}

	showChapterSelector := func() {
		saveCurrentChapter()
		list := tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true)
		list.SetSelectedBackgroundColor(tview.Styles.TitleColor).SetSelectedTextColor(tview.Styles.PrimitiveBackgroundColor)
		list.SetBorder(true).SetTitle("Chapters (< & > reorder)")
		list.SetBorderPadding(1, 1, 2, 2) // Added Padding as requested

		populateList := func() {
			list.Clear()
			for i, chap := range chapters {
				idx := i
				title := fmt.Sprintf("%d. %s", i+1, chap.Title)
				if i == currentChapterIndex {
					title += " (Current)"
				}
				list.AddItem(title, "", 0, func() { loadChapter(idx) })
			}
		}
		populateList()
		list.SetCurrentItem(currentChapterIndex)

		list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEscape {
				pages.HidePage("modal")
				// Restore view focus
				if currentView == ViewNotes {
					app.SetFocus(notesArea)
				} else {
					app.SetFocus(textArea)
				}
				return nil
			}
			if event.Key() == tcell.KeyRune {
				currentItemIndex := list.GetCurrentItem()
				moved := false
				if event.Rune() == '<' && currentItemIndex > 0 {
					chapters[currentItemIndex], chapters[currentItemIndex-1] = chapters[currentItemIndex-1], chapters[currentItemIndex]
					if currentChapterIndex == currentItemIndex {
						currentChapterIndex--
					} else if currentChapterIndex == currentItemIndex-1 {
						currentChapterIndex++
					}
					list.SetCurrentItem(currentItemIndex - 1)
					moved = true
				} else if event.Rune() == '>' && currentItemIndex < len(chapters)-1 {
					chapters[currentItemIndex], chapters[currentItemIndex+1] = chapters[currentItemIndex+1], chapters[currentItemIndex]
					if currentChapterIndex == currentItemIndex {
						currentChapterIndex++
					} else if currentChapterIndex == currentItemIndex+1 {
						currentChapterIndex--
					}
					list.SetCurrentItem(currentItemIndex + 1)
					moved = true
				}
				if moved {
					populateList()
				}
				return nil
			}
			return event
		})

		grid := tview.NewGrid().SetColumns(0, 40, 0).SetRows(0, 20, 0).AddItem(list, 1, 1, 1, 1, 0, 0, true)
		pages.AddPage("modal", grid, true, true)
		app.SetFocus(list)
	}

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
		case "wordcount":
			targetArea := textArea
			if currentView == ViewNotes {
				targetArea = notesArea
			}
			text := targetArea.GetText()
			words := len(strings.Fields(text))
			lines := strings.Count(text, "\n") + 1
			if len(text) == 0 {
				lines = 0
			}
			showModal("Stats", fmt.Sprintf("Words: %d\nChars: %d\nLines: %d", words, len(text), lines))
		case "chapters", "list":
			showChapterSelector()
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
				}
				count := strings.Count(targetArea.GetText(), term)
				showModal("Search", fmt.Sprintf("Found %d of '%s'", count, term))
			}
		case "replace":
			if len(parts) == 3 {
				targetArea := textArea
				if currentView == ViewNotes {
					targetArea = notesArea
				}
				oldT, newT := parts[1], parts[2]
				newText := strings.ReplaceAll(targetArea.GetText(), oldT, newT)
				targetArea.SetText(newText, false)
				showModal("Replace", fmt.Sprintf("Replaced '%s' with '%s'", oldT, newT))
			}
		case "target":
			if len(parts) > 1 {
				if n, err := strconv.Atoi(parts[1]); err == nil {
					setChapterTarget(currentChapterIndex, n)
				}
			} else {
				setChapterTarget(currentChapterIndex, 0)
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
					idx := currentChapterIndex
					nameStart := 2
					if len(parts) > 2 {
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
		targetArea := textArea
		if currentView == ViewNotes {
			targetArea = notesArea
		}
		if currentView == ViewAnalyze {
			position.SetText(" Read-Only ")
			return
		}

		fromRow, fromColumn, _, _ := targetArea.GetCursor()
		text := targetArea.GetText()
		wordCount := len(strings.Fields(text))

		wordCountStr := fmt.Sprintf("[%s]%d[white]", tview.Styles.SecondaryTextColor, wordCount)

		if currentView == ViewMain && chapters[currentChapterIndex].Target > 0 {
			tgt := chapters[currentChapterIndex].Target
			perc := 0
			if wordCount > 0 {
				perc = int((float64(wordCount) / float64(tgt)) * 100)
			}
			col := "white"
			if perc >= 100 {
				col = "green"
			}
			wordCountStr = fmt.Sprintf("%d / %d ([%s]%d%%[-])", wordCount, tgt, col, perc)
		}
		position.SetText(fmt.Sprintf("Words: %s | Row: %d Col: %d ", wordCountStr, fromRow, fromColumn))
	}
	textArea.SetMovedFunc(updateInfos)
	notesArea.SetMovedFunc(updateInfos)
	updateInfos()

	commandPalette.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			cmd := commandPalette.GetText()
			commandPalette.SetText("")
			handleCommand(cmd)

			// Intelligent focus restoration
			isModal := false
			// FIX: ADDED "load", "target", "chapter" to modal list to prevent focus stealing
			for _, m := range []string{"help", "chapters", "list", "wordcount", "save", "open", "load", "export", "search", "replace", "spell", "theme", "analyze", "target", "chapter"} {
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
			} else {
				app.SetFocus(textArea)
			}
		}
	})

	// --- 5. HELP ---
	help1 := tview.NewTextView().SetDynamicColors(true).SetText(`[green]Navigation
[yellow]Arrows[white]: Move cursor
[yellow]Ctrl-A/Home[white]: Start of line
[yellow]Ctrl-E/End[white]: End of line
[blue]Enter for next page, Esc to return.`)

	help2 := tview.NewTextView().SetDynamicColors(true).SetText(`[green]Editing & View
Type to enter text.
[yellow]Ctrl-Q[white]: Copy | [yellow]Ctrl-X[white]: Cut | [yellow]Ctrl-V[white]: Paste
[yellow]Ctrl-Z[white]: Undo | [yellow]Ctrl-Y[white]: Redo
[yellow]Ctrl-T[white]: Toggle Center View
[blue]Enter for next page, Esc to return.`)

	helpCmds := tview.NewTextView().SetDynamicColors(true).SetText(`[green]Commands (Ctrl-E)
[yellow]save/open/export <file>[white]: File ops
[yellow]notes[white] (or Ctrl-N): Toggle Notes
[yellow]analyze[white]: Hemingway Analysis Mode
[yellow]theme <light|dark|retro>[white]: Switch theme
[yellow]target <N>[white]: Set word goal
[yellow]spellcheck[white]: Run check
[yellow]search/replace[white]: Find/Replace
[yellow]chapters[white]: Manage chapters
[yellow]chapter new/delete/rename[white]
[blue]Enter for next page, Esc to return.`)

	// Setup the frame for Help pages
	help := tview.NewFrame(help1).SetBorders(1, 1, 0, 0, 2, 2)
	help.SetTitle("Help")

	// State tracking for help pagination
	helpPageIndex := 0

	help.SetBorder(true).SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
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

	pages.AddAndSwitchToPage("main", mainView, true).AddPage("help", tview.NewGrid().SetColumns(0, 64, 0).SetRows(0, 22, 0).AddItem(help, 1, 1, 1, 1, 0, 0, true), true, false)

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
		if e.Key() == tcell.KeyCtrlE {
			if app.GetFocus() != commandPalette {
				app.SetFocus(commandPalette)
			} else {
				if currentView == ViewNotes {
					app.SetFocus(notesArea)
				} else if currentView == ViewAnalyze {
					app.SetFocus(analysisView)
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
