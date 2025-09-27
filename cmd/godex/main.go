package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"godex/internal/history"

	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/responses"
)

func main() {
	historyPath, err := history.Locate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "locate history file: %v\n", err)
		os.Exit(1)
	}

	// content, err := history.Read(historyPath)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "%v\n", err)
	// 	os.Exit(1)
	// }
	// fmt.Print(string(content))

	commands, err := history.LatestCommands(historyPath, 10)
	if err != nil {
		fmt.Fprintf(os.Stderr, "derive recent commands: %v\n", err)
		os.Exit(1)
	}

	if len(commands) == 0 {
		return
	}

	ctx := context.Background()
	client := openai.NewClient()

	intentPrompt := buildIntentPrompt(commands)

	resp, err := client.Responses.New(ctx, responses.ResponseNewParams{
		Model: openai.ChatModelGPT4oMini,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(intentPrompt),
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "request command intent: %v\n", err)
		os.Exit(1)
	}

	analysis := strings.TrimSpace(resp.OutputText())
	if analysis == "" {
		return
	}

	fmt.Printf("\n\nPotential intent behind recent commands:\n%s\n", analysis)
}

func buildIntentPrompt(commands []string) string {
	var builder strings.Builder
	builder.WriteString("You are an assistant that infers the likely intent behind a sequence of shell commands. Summarize the user's probable goals succinctly.\n\n")
	builder.WriteString("Commands (oldest to newest):\n")
	for idx, cmd := range commands {
		builder.WriteString(fmt.Sprintf("%d. %s\n", idx+1, cmd))
	}
	builder.WriteString("\nRespond with a brief analysis of what the user was trying to accomplish. Focus on intent, not a step-by-step replay.")
	return builder.String()
}
