package browser

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type EntryType int

const (
	EntryFile EntryType = iota
	EntryDir
	EntryParent
)

type Entry struct {
	Name string
	Path string
	Type EntryType
}

type Browser struct {
	Root     string
	Path     string
	Entries  []Entry
	Selected int
	onSelect func(string)
	onCancel func()
	modal    *tview.Grid
	list     *tview.List
	pathView *tview.TextView
}

func New(root string, onSelect func(string), onCancel func()) *Browser {
	if root == "" {
		root = "."
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		absRoot = root
	}
	return &Browser{
		Root:     absRoot,
		Path:     absRoot,
		Entries:  []Entry{},
		Selected: 0,
		onSelect: onSelect,
		onCancel: onCancel,
	}
}

func (b *Browser) ReadDir(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return err
	}

	b.Path = absPath
	b.Entries = []Entry{}
	b.Selected = 0

	if absPath != "/" {
		b.Entries = append(b.Entries, Entry{
			Name: "..",
			Path: filepath.Dir(absPath),
			Type: EntryParent,
		})
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		entryType := EntryFile
		if entry.IsDir() {
			entryType = EntryDir
		}

		b.Entries = append(b.Entries, Entry{
			Name: entry.Name(),
			Path: filepath.Join(absPath, entry.Name()),
			Type: entryType,
		})
	}

	sort.Slice(b.Entries, func(i, j int) bool {
		if b.Entries[i].Type != b.Entries[j].Type {
			return b.Entries[i].Type > b.Entries[j].Type
		}
		return strings.ToLower(b.Entries[i].Name) < strings.ToLower(b.Entries[j].Name)
	})

	return nil
}

func (b *Browser) Show(app *tview.Application, pages *tview.Pages) {
	if err := b.ReadDir(b.Path); err != nil {
		b.onCancel()
		return
	}

	b.list = tview.NewList()
	b.list.ShowSecondaryText(false)
	b.list.SetHighlightFullLine(true)
	b.list.SetSelectedBackgroundColor(tview.Styles.TitleColor)
	b.list.SetSelectedTextColor(tview.Styles.PrimitiveBackgroundColor)
	b.list.SetBorder(true)
	b.list.SetTitle("File Browser")
	b.list.SetBorderPadding(1, 1, 2, 2)

	for _, entry := range b.Entries {
		displayName := entry.Name
		secondary := ""

		switch entry.Type {
		case EntryParent:
			displayName = "./.  (parent directory)"
			secondary = "[dim]"
		case EntryDir:
			displayName = entry.Name + "/"
			secondary = "[dim][DIR][/dim]"
		default:
			ext := filepath.Ext(entry.Name)
			if ext != "" {
				secondary = "[dim][" + strings.TrimPrefix(ext, ".") + "][/dim]"
			}
		}

		b.list.AddItem(displayName, secondary, 0, func() {
			b.handleSelection(app, pages)
		})
	}

	if b.Selected > 0 && b.Selected < len(b.Entries) {
		b.list.SetCurrentItem(b.Selected)
	}

	b.pathView = tview.NewTextView()
	b.pathView.SetText(b.Path)
	b.pathView.SetTextColor(tview.Styles.PrimaryTextColor)
	b.pathView.SetBorder(false)
	b.pathView.SetDynamicColors(true)

	helpView := tview.NewTextView()
	helpView.SetText("[dim]↑↓ navigate  ·  Enter open/enter  ·  Esc cancel[/dim]")
	helpView.SetTextColor(tcell.ColorGray)
	helpView.SetBorder(false)
	helpView.SetDynamicColors(true)

	grid := tview.NewGrid().
		SetColumns(0, 60, 0).
		SetRows(0, 3, 3, 20, 1, 0).
		AddItem(b.pathView, 1, 1, 1, 1, 0, 0, false).
		AddItem(b.list, 3, 1, 1, 1, 0, 0, true).
		AddItem(helpView, 5, 1, 1, 1, 0, 0, false)

	grid.SetBorder(true)
	grid.SetTitle(" [ File Browser ] ")

	b.modal = tview.NewGrid().
		SetColumns(0, 64, 0).
		SetRows(0, 22, 0).
		AddItem(grid, 1, 1, 1, 1, 0, 0, true)

	b.list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			b.onCancel()
			return nil
		case tcell.KeyEnter:
			b.Selected = b.list.GetCurrentItem()
			b.handleSelection(app, pages)
			return nil
		}
		return event
	})

	pages.AddPage("browser", b.modal, true, true)
	app.SetFocus(b.list)
}

func (b *Browser) handleSelection(app *tview.Application, pages *tview.Pages) {
	b.Selected = b.list.GetCurrentItem()
	if b.Selected < 0 || b.Selected >= len(b.Entries) {
		return
	}

	entry := b.Entries[b.Selected]

	switch entry.Type {
	case EntryParent:
		b.Path = entry.Path
		pages.RemovePage("browser")
		b.Show(app, pages)
	case EntryDir:
		b.Path = entry.Path
		pages.RemovePage("browser")
		b.Show(app, pages)
	case EntryFile:
		pages.HidePage("browser")
		pages.RemovePage("browser")
		if b.onSelect != nil {
			b.onSelect(entry.Path)
		}
	}
}

func (b *Browser) Hide(pages *tview.Pages) {
	if pages != nil {
		pages.HidePage("browser")
		pages.RemovePage("browser")
	}
}
