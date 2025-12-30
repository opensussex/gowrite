package state

import (
	"errors"
	"strings"
	"sync"
)

// Chapter represents a section of the document.
type Chapter struct {
	Title   string
	Content string
	Notes   string
	Target  int
}

// WikiEntry represents a single item in the Story Wiki.
type WikiEntry struct {
	Title   string
	Content string
}

// Project holds chapters and wiki entries for persistence.
type Project struct {
	Chapters []Chapter
	Wiki     []WikiEntry
}

// AppState captures mutable application state shared across UI and logic.
type AppState struct {
	mu sync.Mutex

	Project Project

	CurrentChapterIdx int
	CurrentWikiIdx    int
	Filename          string
	View              int

	IsCentered  bool
	IsFocusMode bool
	IsLoading   bool
}

// CurrentChapterIndex returns the active chapter index.
func (s *AppState) CurrentChapterIndex() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.CurrentChapterIdx
}

// CurrentWikiIndex returns the active wiki index.
func (s *AppState) CurrentWikiIndex() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.CurrentWikiIdx
}

// CurrentFilename returns the last-used filename.
func (s *AppState) CurrentFilename() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Filename
}

// SetCurrentFilename updates the last-used filename.
func (s *AppState) SetCurrentFilename(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Filename = name
}

// CurrentView returns the active UI view.
func (s *AppState) CurrentView() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.View
}

// SetView updates the active UI view.
func (s *AppState) SetView(view int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.View = view
}

// FocusMode returns whether focus mode is enabled.
func (s *AppState) FocusMode() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.IsFocusMode
}

// SetFocusMode explicitly sets focus mode.
func (s *AppState) SetFocusMode(on bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.IsFocusMode = on
}

// Centered returns whether centered layout is enabled.
func (s *AppState) Centered() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.IsCentered
}

// SetCentered toggles centered layout.
func (s *AppState) SetCentered(on bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.IsCentered = on
}

// Loading returns whether a load operation is in progress.
func (s *AppState) Loading() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.IsLoading
}

// SetLoading toggles loading state (suppresses autosave when true).
func (s *AppState) SetLoading(on bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.IsLoading = on
}

// NewAppState builds an AppState with one default chapter and wiki entry.
func NewAppState() *AppState {
	return &AppState{
		Project: Project{
			Chapters: []Chapter{{Title: "The Beginning"}},
			Wiki:     []WikiEntry{{Title: "General Notes"}},
		},
		CurrentChapterIdx: 0,
		CurrentWikiIdx:    0,
		View:              0,
	}
}

// Snapshot returns a shallow copy of the project for safe use outside locks.
func (s *AppState) Snapshot() Project {
	s.mu.Lock()
	defer s.mu.Unlock()

	proj := Project{}
	proj.Chapters = append(proj.Chapters, s.Project.Chapters...)
	proj.Wiki = append(proj.Wiki, s.Project.Wiki...)
	return proj
}

// SetProject replaces the current project and resets indices safely.
func (s *AppState) SetProject(p Project) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Project = p
	if len(s.Project.Wiki) == 0 {
		s.Project.Wiki = []WikiEntry{{Title: "General", Content: ""}}
	}
	s.CurrentChapterIdx = 0
	s.CurrentWikiIdx = 0
}

// SaveCurrentChapter writes content/notes to the current chapter.
func (s *AppState) SaveCurrentChapter(content, notes string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.CurrentChapterIdx < 0 || s.CurrentChapterIdx >= len(s.Project.Chapters) {
		return
	}
	s.Project.Chapters[s.CurrentChapterIdx].Content = content
	s.Project.Chapters[s.CurrentChapterIdx].Notes = notes
}

// SaveCurrentWiki writes content to the current wiki entry.
func (s *AppState) SaveCurrentWiki(content string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.CurrentWikiIdx < 0 || s.CurrentWikiIdx >= len(s.Project.Wiki) {
		return
	}
	s.Project.Wiki[s.CurrentWikiIdx].Content = content
}

// LoadChapter returns the requested chapter by index.
func (s *AppState) LoadChapter(idx int) (Chapter, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if idx < 0 || idx >= len(s.Project.Chapters) {
		return Chapter{}, errors.New("chapter index out of range")
	}
	s.CurrentChapterIdx = idx
	return s.Project.Chapters[idx], nil
}

// LoadWiki returns the requested wiki entry by index.
func (s *AppState) LoadWiki(idx int) (WikiEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if idx < 0 || idx >= len(s.Project.Wiki) {
		return WikiEntry{}, errors.New("wiki index out of range")
	}
	s.CurrentWikiIdx = idx
	return s.Project.Wiki[idx], nil
}

// ApplyStructure replaces chapters with a named template.
func (s *AppState) ApplyStructure(name string) error {
	templates := map[string][]Chapter{
		"3act": {
			{Title: "Act 1: The Setup", Notes: "Introduce characters and the ordinary world.\nEstablish the status quo and the flaw that holds them back.", Content: ">> GUIDANCE: Introduce the protagonist in their 'Ordinary World'. Establish the status quo and the flaw that holds them back."},
			{Title: "Inciting Incident", Notes: "Something happens that disrupts the status quo.\nThe hero faces a problem they cannot ignore.", Content: ">> GUIDANCE: An external event disrupts the status quo. The hero faces a problem they cannot ignore."},
			{Title: "Plot Point 1", Notes: "The hero leaves the ordinary world.\nThe hero decides to engage with the problem.", Content: ">> GUIDANCE: The hero decides to engage with the problem. They leave their comfort zone and cross into the 'Special World'."},
			{Title: "Act 2: The Confrontation", Notes: "Rising action, tests, allies, and enemies.", Content: ">> GUIDANCE: Rising action. The hero meets allies and enemies. They face tests that force them to learn new skills."},
			{Title: "Midpoint", Notes: "A major event shifts the context (false victory/defeat).\nThe stakes are raised; there is no turning back.", Content: ">> GUIDANCE: A major event shifts the context (a false victory or defeat). The stakes are raised; there is no turning back."},
			{Title: "Plot Point 2", Notes: "All hope seems lost (The Dark Night of the Soul).\nThe hero must find a new solution or inner strength.", Content: ">> GUIDANCE: All hope seems lost. The hero must find a new solution or inner strength."},
			{Title: "Act 3: The Resolution", Notes: "The final battle/climax.\nThe hero faces the antagonist one last time.", Content: ">> GUIDANCE: The Climax. The hero faces the antagonist one last time. They must use the lessons learned in Act 2 to win."},
			{Title: "The End", Notes: "The aftermath. Establish the 'New Normal'.\nShow how the hero has changed.", Content: ">> GUIDANCE: The aftermath. Establish the 'New Normal'. Show how the hero has changed."},
		},
		"hero": {
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
		},
		"cat": {
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
		},
		"fichtean": {
			{Title: "Inciting Incident", Notes: "Start immediately with the problem.", Content: ">> GUIDANCE: Skip the setup. Start immediately with the problem. Throw the reader into the action."},
			{Title: "Crisis 1", Notes: "First obstacle. Rising action.", Content: ">> GUIDANCE: The first major obstacle. The hero tries to solve it but complications arise."},
			{Title: "Crisis 2", Notes: "Higher stakes obstacle.", Content: ">> GUIDANCE: The stakes get higher. The problem expands or gets more personal."},
			{Title: "Crisis 3", Notes: "Even higher stakes.", Content: ">> GUIDANCE: The situation seems dire. The hero's resources are running thin."},
			{Title: "The Climax", Notes: "Maximum tension.", Content: ">> GUIDANCE: Maximum tension. The final confrontation. The hero succeeds or fails."},
			{Title: "Falling Action", Notes: "Loose ends tied.", Content: ">> GUIDANCE: Loose ends are tied up. The immediate aftermath of the climax."},
			{Title: "Resolution", Notes: "New normal.", Content: ">> GUIDANCE: The new normal is established. A brief moment of calm."},
		},
		"horror": {
			{Title: "The Dreadful Normal", Notes: "Establish status quo with unease.", Content: ">> GUIDANCE: Establish the setting and characters. Create a subtle sense of unease or isolation despite the normalcy."},
			{Title: "The Omen", Notes: "A warning sign.", Content: ">> GUIDANCE: A warning sign appears but is ignored or rationalized. The first subtle brush with the entity."},
			{Title: "The Onset", Notes: "The threat reveals itself.", Content: ">> GUIDANCE: The threat reveals itself properly. The first scare or victim. There is no going back now."},
			{Title: "The Discovery", Notes: "Realization of the horror.", Content: ">> GUIDANCE: The characters realize what they are dealing with. Escape attempts fail. Isolation is complete."},
			{Title: "The Pursuit", Notes: "Cat and Mouse.", Content: ">> GUIDANCE: The entity attacks. High tension chase or siege. The characters are stripped of resources."},
			{Title: "The Confrontation", Notes: "The final stand.", Content: ">> GUIDANCE: The final stand. The remaining survivors must face the horror head-on. High casualty rate."},
			{Title: "The Aftermath", Notes: "Survival... or is it?", Content: ">> GUIDANCE: The evil is defeated... or is it? The survivors escape, but they are changed forever."},
		},
	}

	key := strings.ToLower(name)
	tpl, ok := templates[key]
	if !ok {
		return errors.New("unknown structure")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.Project.Chapters = append([]Chapter(nil), tpl...)
	if len(s.Project.Wiki) == 0 {
		s.Project.Wiki = []WikiEntry{{Title: "General", Content: ""}}
	}
	s.CurrentChapterIdx = 0
	s.CurrentWikiIdx = 0
	return nil
}

// ToggleView switches current view and returns the new view.
func (s *AppState) ToggleView(view int) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.View = view
	return s.View
}

// ToggleFocus flips focus mode and returns the new state.
func (s *AppState) ToggleFocus() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.IsFocusMode = !s.IsFocusMode
	return s.IsFocusMode
}

// ToggleNotesView convenience wrapper for view management.
func (s *AppState) ToggleNotesView(notesView int, mainView int) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.View == notesView {
		s.View = mainView
	} else {
		s.View = notesView
	}
	return s.View
}

// ToggleWikiView convenience wrapper for view management.
func (s *AppState) ToggleWikiView(wikiView int, mainView int) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.View == wikiView {
		s.View = mainView
	} else {
		s.View = wikiView
	}
	return s.View
}

// NewChapter appends a new chapter and returns its index.
func (s *AppState) NewChapter(title string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Project.Chapters = append(s.Project.Chapters, Chapter{Title: title})
	s.CurrentChapterIdx = len(s.Project.Chapters) - 1
	return s.CurrentChapterIdx
}

// RenameChapter renames a chapter at index.
func (s *AppState) RenameChapter(idx int, title string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if idx < 0 || idx >= len(s.Project.Chapters) {
		return errors.New("chapter index out of range")
	}
	s.Project.Chapters[idx].Title = title
	return nil
}

// DeleteChapter removes a chapter and returns the new current index.
func (s *AppState) DeleteChapter(idx int) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.Project.Chapters) <= 1 {
		return s.CurrentChapterIdx, errors.New("cannot delete only chapter")
	}
	if idx < 0 || idx >= len(s.Project.Chapters) {
		return s.CurrentChapterIdx, errors.New("chapter index out of range")
	}
	s.Project.Chapters = append(s.Project.Chapters[:idx], s.Project.Chapters[idx+1:]...)
	if idx <= s.CurrentChapterIdx && s.CurrentChapterIdx > 0 {
		s.CurrentChapterIdx--
	}
	if s.CurrentChapterIdx >= len(s.Project.Chapters) {
		s.CurrentChapterIdx = len(s.Project.Chapters) - 1
	}
	return s.CurrentChapterIdx, nil
}

// NewWiki appends a new wiki entry and returns its index.
func (s *AppState) NewWiki(title string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Project.Wiki = append(s.Project.Wiki, WikiEntry{Title: title})
	s.CurrentWikiIdx = len(s.Project.Wiki) - 1
	return s.CurrentWikiIdx
}

// RenameWiki renames a wiki entry at index.
func (s *AppState) RenameWiki(idx int, title string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if idx < 0 || idx >= len(s.Project.Wiki) {
		return errors.New("wiki index out of range")
	}
	s.Project.Wiki[idx].Title = title
	return nil
}

// DeleteWiki removes a wiki entry and returns the new current index.
func (s *AppState) DeleteWiki(idx int) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.Project.Wiki) <= 1 {
		return s.CurrentWikiIdx, errors.New("cannot delete only wiki entry")
	}
	if idx < 0 || idx >= len(s.Project.Wiki) {
		return s.CurrentWikiIdx, errors.New("wiki index out of range")
	}
	s.Project.Wiki = append(s.Project.Wiki[:idx], s.Project.Wiki[idx+1:]...)
	if idx <= s.CurrentWikiIdx && s.CurrentWikiIdx > 0 {
		s.CurrentWikiIdx--
	}
	if s.CurrentWikiIdx >= len(s.Project.Wiki) {
		s.CurrentWikiIdx = len(s.Project.Wiki) - 1
	}
	return s.CurrentWikiIdx, nil
}
