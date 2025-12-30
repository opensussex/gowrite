package persistence

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"gowrite/state"
)

// SaveProject writes the project to disk as JSON.
func SaveProject(path string, p state.Project) error {
	if path == "" {
		return errors.New("filename required")
	}
	if filepath.Ext(path) == "" {
		path += ".json"
	}

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadProject reads a project from disk, supporting new and old formats.
func LoadProject(path string) (state.Project, error) {
	if path == "" {
		return state.Project{}, errors.New("filename required")
	}
	if filepath.Ext(path) == "" {
		path += ".json"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return state.Project{}, err
	}

	// Try new format first.
	var project state.Project
	if err := json.Unmarshal(data, &project); err == nil && len(project.Chapters) > 0 {
		if len(project.Wiki) == 0 {
			project.Wiki = []state.WikiEntry{{Title: "General", Content: ""}}
		}
		return project, nil
	}

	// Fallback: old format was []Chapter only.
	var oldChapters []state.Chapter
	if err := json.Unmarshal(data, &oldChapters); err == nil && len(oldChapters) > 0 {
		return state.Project{Chapters: oldChapters, Wiki: []state.WikiEntry{{Title: "General", Content: ""}}}, nil
	}

	return state.Project{}, errors.New("file empty or corrupt")
}

// ExportText writes chapters as plain text with headings.
func ExportText(path string, p state.Project) error {
	if path == "" {
		return errors.New("filename required")
	}
	if filepath.Ext(path) == "" {
		path += ".txt"
	}

	var buf []byte
	for i, chap := range p.Chapters {
		heading := []byte("# Chapter " + itoa(i+1) + ": " + chap.Title + "\n\n")
		buf = append(buf, heading...)
		buf = append(buf, []byte(chap.Content)...)
		buf = append(buf, []byte("\n\n")...)
	}
	return os.WriteFile(path, buf, 0644)
}

// itoa avoids pulling strconv for this small helper.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		d := byte('0' + n%10)
		digits = append([]byte{d}, digits...)
		n /= 10
	}
	return string(digits)
}
