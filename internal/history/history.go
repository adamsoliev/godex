package history

import (
	"fmt"
	"os"
	"path/filepath"
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
