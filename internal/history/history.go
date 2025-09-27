package history

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	lines := make([]string, 0, count)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan history file %s: %w", historyPath, err)
	}

	commands := make([]string, 0, count)

	for i := len(lines) - 1; i >= 0 && len(commands) < count; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, ": ") {
			if idx := strings.Index(line, ";"); idx != -1 && idx+1 < len(line) {
				line = line[idx+1:]
			}
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		commands = append(commands, line)
	}

	// Reverse to chronological order (oldest first)
	for i, j := 0, len(commands)-1; i < j; i, j = i+1, j-1 {
		commands[i], commands[j] = commands[j], commands[i]
	}

	return commands, nil
}
