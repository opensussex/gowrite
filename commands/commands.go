package commands

import (
	"errors"
	"strings"

	"gowrite/persistence"
	"gowrite/state"
)

// Result represents the outcome of a command execution.
type Result struct {
	Message string
	Err     error
	Modal   bool
}

// Handler processes a command with arguments against the shared state.
type Handler func(args []string, st *state.AppState) Result

// Registry maps command names to handlers.
var Registry = map[string]Handler{}

// Register built-in handlers.
func init() {
	Registry["save"] = saveHandler
	Registry["open"] = openHandler
	Registry["load"] = openHandler
	Registry["export"] = exportHandler
}

func saveHandler(args []string, st *state.AppState) Result {
	filename := strings.TrimSpace(strings.Join(args, " "))
	if filename == "" {
		filename = st.CurrentFilename()
	}
	if filename == "" {
		return Result{Err: errors.New("please provide a filename")}
	}

	proj := st.Snapshot()
	if err := persistence.SaveProject(filename, proj); err != nil {
		return Result{Err: err}
	}
	st.SetCurrentFilename(filename)
	return Result{Message: "Saved to " + filename, Modal: true}
}

func openHandler(args []string, st *state.AppState) Result {
	filename := strings.TrimSpace(strings.Join(args, " "))
	if filename == "" {
		return Result{Err: errors.New("filename required")}
	}

	proj, err := persistence.LoadProject(filename)
	if err != nil {
		return Result{Err: err}
	}
	st.SetProject(proj)
	st.SetCurrentFilename(filename)
	return Result{Message: "Loaded " + filename, Modal: true}
}

func exportHandler(args []string, st *state.AppState) Result {
	if len(args) == 0 {
		return Result{Err: errors.New("usage: export <file>")}
	}
	filename := strings.TrimSpace(strings.Join(args, " "))
	if filename == "" {
		return Result{Err: errors.New("filename required")}
	}
	proj := st.Snapshot()
	if err := persistence.ExportText(filename, proj); err != nil {
		return Result{Err: err}
	}
	return Result{Message: "Exported to " + filename, Modal: true}
}
