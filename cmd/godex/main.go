package main

import (
	"fmt"
	"os"

	"godex/internal/history"
)

func main() {
	historyPath, err := history.Locate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "locate history file: %v\n", err)
		os.Exit(1)
	}

	content, err := history.Read(historyPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	fmt.Print(string(content))
}
