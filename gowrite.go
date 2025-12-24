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

	chapters := []Chapter{
		{Title: "The Beginning", Content: "", Notes: "", Target: 0},
	}

	wikiEntries := []WikiEntry{
		{Title: "General Notes", Content: ""},
	}

	currentChapterIndex := 0
	currentWikiIndex := 0
	currentFilename := ""
	currentView := ViewMain

	// Visual States
	isCenteredView := false
	isFocusMode := false // Hides all UI chrome

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

	defaultHelpText := " F1: Help | Ctrl-G: Chapters | Ctrl-N: Notes | Ctrl-W: Wiki | Ctrl-E: Command"
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

		wikiList.Clear()
		for i, w := range wikiEntries {
			title := w.Title
			if i == currentWikiIndex {
				title += " *"
			}
			idx := i
			wikiList.AddItem(title, "", 0, func() {
				loadWiki(idx)
				app.SetFocus(wikiArea)
			})
		}
		wikiList.SetCurrentItem(currentWikiIndex)
	}

	setView := func(viewType int) {
		if currentView == ViewWiki {
			saveCurrentWiki()
		} else {
			saveCurrentChapter()
		}

		currentView = viewType
		mainView.Clear()

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

			mainView.SetColumns(30, 0)
			mainView.SetRows(0, 3, 1)

			mainView.AddItem(wikiList, 0, 0, 1, 1, 0, 0, true)
			mainView.AddItem(wikiArea, 0, 1, 1, 1, 0, 0, false)
			mainView.AddItem(commandPalette, 1, 0, 1, 2, 0, 0, false)
			mainView.AddItem(helpInfo, 2, 0, 1, 1, 0, 0, false)
			mainView.AddItem(position, 2, 1, 1, 1, 0, 0, false)

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
		setView(currentView)
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

	// --- STRUCTURE TEMPLATES ---
	applyStructure := func(name string) {
		var newChapters []Chapter
		name = strings.ToLower(name)

		switch name {
		case "3act", "standard":
			newChapters = []Chapter{
				{Title: "Act 1: The Setup", Notes: "Introduce characters and the ordinary world.\nEstablish the status quo and the flaw that holds them back.", Content: ">> GUIDANCE: Introduce the protagonist in their 'Ordinary World'. Establish the status quo and the flaw that holds them back."},
				{Title: "Inciting Incident", Notes: "Something happens that disrupts the status quo.\nThe hero faces a problem they cannot ignore.", Content: ">> GUIDANCE: An external event disrupts the status quo. The hero faces a problem they cannot ignore."},
				{Title: "Plot Point 1", Notes: "The hero leaves the ordinary world.\nThe hero decides to engage with the problem.", Content: ">> GUIDANCE: The hero decides to engage with the problem. They leave their comfort zone and cross into the 'Special World'."},
				{Title: "Act 2: The Confrontation", Notes: "Rising action, tests, allies, and enemies.", Content: ">> GUIDANCE: Rising action. The hero meets allies and enemies. They face tests that force them to learn new skills."},
				{Title: "Midpoint", Notes: "A major event shifts the context (false victory/defeat).\nThe stakes are raised; there is no turning back.", Content: ">> GUIDANCE: A major event shifts the context (a false victory or defeat). The stakes are raised; there is no turning back."},
				{Title: "Plot Point 2", Notes: "All hope seems lost (The Dark Night of the Soul).\nThe hero must find a new solution or inner strength.", Content: ">> GUIDANCE: All hope seems lost. The hero must find a new solution or inner strength."},
				{Title: "Act 3: The Resolution", Notes: "The final battle/climax.\nThe hero faces the antagonist one last time.", Content: ">> GUIDANCE: The Climax. The hero faces the antagonist one last time. They must use the lessons learned in Act 2 to win."},
				{Title: "The End", Notes: "The aftermath. Establish the 'New Normal'.\nShow how the hero has changed.", Content: ">> GUIDANCE: The aftermath. Establish the 'New Normal'. Show how the hero has changed."},
			}
		case "hero", "monomyth":
			newChapters = []Chapter{
				{Title: "The Ordinary World", Notes: "Status Quo.", Content: ">> GUIDANCE: Show the hero's life before the journey. Highlight their dissatisfaction or lack of completeness."},
				{Title: "Call to Adventure", Notes: "Disruption.", Content: ">> GUIDANCE: Something shakes up the situation. The hero is presented with a challenge or opportunity."},
				{Title: "Refusal of the Call", Notes: "Fear or hesitation.", Content: ">> GUIDANCE: The hero hesitates due to fear or insecurity. Why are they afraid to leave?"},
				{Title: "Meeting the Mentor", Notes: "Gaining tools/advice.", Content: ">> GUIDANCE: The hero gains supplies, advice, or confidence from a mentor. They are now ready to face the journey."},
				{Title: "Crossing the Threshold", Notes: "Leaving the known world.", Content: ">> GUIDANCE: The hero commits to leaving the Ordinary World. They enter the Special World with different rules."},
				{Title: "Tests, Allies, Enemies", Notes: "Learning the rules.", Content: ">> GUIDANCE: The hero explores the new world. They make friends and attract enemies."},
				{Title: "Approach to the Cave", Notes: "Preparing for the main danger.", Content: ">> GUIDANCE: The hero prepares for the major challenge. Plans are made, and the team is gathered."},
				{Title: "The Ordeal", Notes: "Death and rebirth moment.", Content: ">> GUIDANCE: The central crisis (midpoint). A brush with death. The hero confronts their greatest fear."},
				{Title: "The Reward", Notes: "Seizing the sword.", Content: ">> GUIDANCE: The hero seizes the object of their quest (sword, elixir, knowledge). But the danger is not over yet."},
				{Title: "The Road Back", Notes: "The chase scene/urgency.", Content: ">> GUIDANCE: The hero is pursued by the vengeful forces. The urgency ramps up for the final escape."},
				{Title: "Resurrection", Notes: "Final test.", Content: ">> GUIDANCE: The final test. The hero is purified by a last sacrifice. They must prove they have truly learned the lesson."},
				{Title: "Return with Elixir", Notes: "Master of two worlds.", Content: ">> GUIDANCE: The hero returns home, transformed. They bring back something that heals the Ordinary World."},
			}
		case "cat", "save the cat":
			newChapters = []Chapter{
				{Title: "Opening Image", Notes: "Snapshot of life before.", Content: ">> GUIDANCE: A visual snapshot of the status quo. Set the tone and mood."},
				{Title: "Theme Stated", Notes: "What the story is really about.", Content: ">> GUIDANCE: Someone (usually not the hero) states the theme of the story. The hero doesn't understand it yet."},
				{Title: "Setup", Notes: "Expanding on the hero's flaws.", Content: ">> GUIDANCE: Expand on the hero's life and flaws. Show why they need to change (Stasis = Death)."},
				{Title: "Catalyst", Notes: "Life changes forever.", Content: ">> GUIDANCE: The Inciting Incident. Life changes forever; they can't go back."},
				{Title: "Debate", Notes: "Can I do this?", Content: ">> GUIDANCE: The hero reacts to the catalyst. They question what to do (Refusal of the Call)."},
				{Title: "Break into Two", Notes: "Choosing the journey.", Content: ">> GUIDANCE: The hero makes a proactive choice to enter the new world. Act 2 begins."},
				{Title: "B Story", Notes: "Love interest or subplot.", Content: ">> GUIDANCE: Introduce the love interest or subplot character. This relationship discusses the theme."},
				{Title: "Fun and Games", Notes: "The 'trailer' moments.", Content: ">> GUIDANCE: The 'Promise of the Premise'. Show scenes that audiences came to see."},
				{Title: "Midpoint", Notes: "Stakes raise significantly.", Content: ">> GUIDANCE: Stakes raise significantly (False Victory or False Defeat). The 'clock' starts ticking."},
				{Title: "Bad Guys Close In", Notes: "Pressure mounts.", Content: ">> GUIDANCE: Internal and external pressure mounts. The hero's plan starts to fail."},
				{Title: "All Is Lost", Notes: "Whiff of death.", Content: ">> GUIDANCE: The lowest point. Something dies (literally or metaphorically). The hero loses hope."},
				{Title: "Dark Night of the Soul", Notes: "Wallowing in hopelessness.", Content: ">> GUIDANCE: The hero wallows in their hopelessness. But in the darkness, they find the true solution."},
				{Title: "Break into Three", Notes: "The new idea/solution.", Content: ">> GUIDANCE: The hero realizes the answer (fixing the flaw). They devise a new plan."},
				{Title: "Finale", Notes: "Executing the plan.", Content: ">> GUIDANCE: The hero executes the plan and defeats the bad guys. The old world is destroyed/changed."},
				{Title: "Final Image", Notes: "Mirror of opening image.", Content: ">> GUIDANCE: Mirror of the Opening Image. Show visually how much the hero has changed."},
			}
		case "fichtean":
			newChapters = []Chapter{
				{Title: "Inciting Incident", Notes: "Start immediately with the problem.", Content: ">> GUIDANCE: Skip the setup. Start immediately with the problem. Throw the reader into the action."},
				{Title: "Crisis 1", Notes: "First obstacle. Rising action.", Content: ">> GUIDANCE: The first major obstacle. The hero tries to solve it but complications arise."},
				{Title: "Crisis 2", Notes: "Higher stakes obstacle.", Content: ">> GUIDANCE: The stakes get higher. The problem expands or gets more personal."},
				{Title: "Crisis 3", Notes: "Even higher stakes.", Content: ">> GUIDANCE: The situation seems dire. The hero's resources are running thin."},
				{Title: "The Climax", Notes: "Maximum tension.", Content: ">> GUIDANCE: Maximum tension. The final confrontation. The hero succeeds or fails."},
				{Title: "Falling Action", Notes: "Loose ends tied.", Content: ">> GUIDANCE: Loose ends are tied up. The immediate aftermath of the climax."},
				{Title: "Resolution", Notes: "New normal.", Content: ">> GUIDANCE: The new normal is established. A brief moment of calm."},
			}
		case "horror":
			newChapters = []Chapter{
				{Title: "The Dreadful Normal", Notes: "Establish status quo with unease.", Content: ">> GUIDANCE: Establish the setting and characters. Create a subtle sense of unease or isolation despite the normalcy."},
				{Title: "The Omen", Notes: "A warning sign.", Content: ">> GUIDANCE: A warning sign appears but is ignored or rationalized. The first subtle brush with the entity."},
				{Title: "The Onset", Notes: "The threat reveals itself.", Content: ">> GUIDANCE: The threat reveals itself properly. The first scare or victim. There is no going back now."},
				{Title: "The Discovery", Notes: "Realization of the horror.", Content: ">> GUIDANCE: The characters realize what they are dealing with. Escape attempts fail. Isolation is complete."},
				{Title: "The Pursuit", Notes: "Cat and Mouse.", Content: ">> GUIDANCE: The entity attacks. High tension chase or siege. The characters are stripped of resources."},
				{Title: "The Confrontation", Notes: "The final stand.", Content: ">> GUIDANCE: The final stand. The remaining survivors must face the horror head-on. High casualty rate."},
				{Title: "The Aftermath", Notes: "Survival... or is it?", Content: ">> GUIDANCE: The evil is defeated... or is it? The survivors escape, but they are changed forever."},
			}
		default:
			showModal("Error", "Unknown structure.\nTry: 3act, hero, cat, fichtean, horror")
			return
		}

		showYesNoModal("Warning", fmt.Sprintf("This will ERASE all current chapters and apply '%s'. Continue?", name), func() {
			chapters = newChapters
			currentChapterIndex = 0
			// FIX: Manually update UI to avoid 'loadChapter' saving old blank text over new template
			textArea.SetText(chapters[0].Content, false)
			notesArea.SetText(chapters[0].Notes, false)
			textArea.SetTitle(fmt.Sprintf("gowrite - Chapter 1: %s", chapters[0].Title))
			flashStatusMessage("Applied Structure: " + name)
		})
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
				if currentView == ViewNotes {
					app.SetFocus(notesArea)
				} else if currentView == ViewWiki {
					app.SetFocus(wikiArea)
				} else {
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
				showFilePicker()
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
			for _, m := range []string{"help", "chapters", "list", "wordcount", "save", "open", "load", "export", "search", "replace", "spell", "theme", "analyze", "target", "chapter", "wiki", "structure"} {
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
[yellow]structure <type>[white]: Apply template (3act, hero, cat, fichtean, horror)
[yellow]wiki[white]: Open Story Bible (Ctrl-W to close)
[yellow]wiki new <name>[white]: Add entry
[yellow]wiki rename <name>[white]: Rename entry
[yellow]wiki delete[white]: Delete entry
[yellow]save <file>[white]: Save project
[yellow]open[white]: Show file picker (or [yellow]open <file>[white] to open directly)
[yellow]export <file>[white]: Export to text
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
		if e.Key() == tcell.KeyCtrlG {
			handleCommand("chapters")
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
