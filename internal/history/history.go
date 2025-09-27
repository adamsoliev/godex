package history

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Locate returns the path to the user's shell history file.
func Locate() (string, error) {
	if hist := os.Getenv("HISTFILE"); hist != "" {
		if _, err := os.Stat(hist); err == nil {
			return hist, nil
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determine home directory: %w", err)
	}

	candidates := []string{
		filepath.Join(home, ".zsh_history"),
		filepath.Join(home, ".bash_history"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("no known history file found")
}

// Read returns the contents of the history file at path.
func Read(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read history file %s: %w", path, err)
	}

	return data, nil
}

// LatestCommands returns up to count commands from historyPath, ordered oldest to newest.
// If count is zero or negative, an empty slice is returned.
func LatestCommands(historyPath string, count int) ([]string, error) {
	if count <= 0 {
		return []string{}, nil
	}

	file, err := os.Open(historyPath)
	if err != nil {
		return nil, fmt.Errorf("open history file %s: %w", historyPath, err)
	}
	defer file.Close()

	entries := make([]Entry, 0, count)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if entry, ok := parseHistoryLine(scanner.Text(), time.Local); ok {
			entries = append(entries, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan history file %s: %w", historyPath, err)
	}

	commands := make([]string, 0, count)

	for i := len(entries) - 1; i >= 0 && len(commands) < count; i-- {
		entry := entries[i]
		if entry.Command == "" {
			continue
		}
		commands = append(commands, entry.Command)
	}

	// Reverse to chronological order (oldest first)
	for i, j := 0, len(commands)-1; i < j; i, j = i+1, j-1 {
		commands[i], commands[j] = commands[j], commands[i]
	}

	return commands, nil
}

// Entry represents a history line with an optional timestamp and the command text.
type Entry struct {
	Timestamp time.Time
	Command   string
}

// DailyEntries returns commands executed on the provided day (local time).
// Commands are returned in chronological order.
func DailyEntries(historyPath string, day time.Time) ([]Entry, error) {
	file, err := os.Open(historyPath)
	if err != nil {
		return nil, fmt.Errorf("open history file %s: %w", historyPath, err)
	}
	defer file.Close()

	targetYear, targetMonth, targetDay := day.Date()
	location := day.Location()

	entries := make([]Entry, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		entry, ok := parseHistoryLine(scanner.Text(), location)
		if !ok {
			continue
		}
		if entry.Timestamp.IsZero() {
			continue
		}
		year, month, day := entry.Timestamp.Date()
		if year == targetYear && month == targetMonth && day == targetDay {
			entries = append(entries, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan history file %s: %w", historyPath, err)
	}

	return entries, nil
}

func parseHistoryLine(line string, location *time.Location) (Entry, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return Entry{}, false
	}

	if strings.HasPrefix(trimmed, ": ") {
		rest := trimmed[2:]
		semi := strings.Index(rest, ";")
		if semi == -1 || semi+1 >= len(rest) {
			return Entry{}, false
		}

		meta := rest[:semi]
		cmd := strings.TrimSpace(rest[semi+1:])
		if cmd == "" {
			return Entry{}, false
		}

		entry := Entry{Command: cmd}

		if colon := strings.Index(meta, ":"); colon != -1 {
			tsStr := meta[:colon]
			if seconds, err := strconv.ParseInt(tsStr, 10, 64); err == nil {
				ts := time.Unix(seconds, 0)
				if location != nil {
					ts = ts.In(location)
				} else {
					ts = ts.Local()
				}
				entry.Timestamp = ts
			}
		}

		return entry, true
	}

	return Entry{Command: trimmed}, true
}
