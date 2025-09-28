package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"godex/internal/history"

	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/responses"
)

func main() {
	historyPath, err := history.Locate()
	entries, err := history.DailyEntries(historyPath, time.Now())
	if err != nil {
		fmt.Fprintf(os.Stderr, "collect today's commands: %v\n", err)
		os.Exit(1)
	}

	if len(entries) == 0 {
		return
	}

	ctx := context.Background()
	client := openai.NewClient()

	summaryPrompt := buildIntentSummaryPrompt(entries)
	summary, err := requestResponseText(ctx, client, summaryPrompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "summarize intents: %v\n", err)
		os.Exit(1)
	}
	summary = strings.TrimSpace(summary)
	if summary == "" {
		return
	}

	refinementPrompt := buildOptimizationPrompt(summary)
	suggestions, err := requestResponseText(ctx, client, refinementPrompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "optimize workflow: %v\n", err)
		os.Exit(1)
	}
	suggestions = strings.TrimSpace(suggestions)

	webPrompt := buildWebSearchPrompt(summary, suggestions)
	webFindings, err := requestResponseText(ctx, client, webPrompt, withWebSearchTool())
	if err != nil {
		fmt.Fprintf(os.Stderr, "research with web search: %v\n", err)
		os.Exit(1)
	}
	webFindings = strings.TrimSpace(webFindings)

	fmt.Printf("\n\nToday's intent summary:\n%s\n", summary)
	if suggestions != "" {
		fmt.Printf("\n\nImprovement ideas:\n%s\n", suggestions)
	}
	if webFindings != "" {
		fmt.Printf("\n\nWeb discoveries:\n%s\n", webFindings)
	}
}

func requestResponseText(ctx context.Context, client openai.Client, prompt string, configure ...func(*responses.ResponseNewParams)) (string, error) {
	params := responses.ResponseNewParams{
		Model: openai.ChatModelGPT4oMini,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(prompt),
		},
	}

	for _, fn := range configure {
		fn(&params)
	}

	resp, err := client.Responses.New(ctx, params)
	if err != nil {
		return "", err
	}

	return resp.OutputText(), nil
}

func buildIntentSummaryPrompt(entries []history.Entry) string {
	var builder strings.Builder
	builder.WriteString("You are observing a user's shell activity. Summarize, in at least three concise bullet points, the primary intents these commands suggest. Focus on the user's high-level goals. For each bullet, append parentheses that cite representative commands, generalized to their base command (e.g., `git status` -> `git`).\n\n")
	builder.WriteString("Today's commands (chronological):\n")
	for _, entry := range entries {
		timestamp := entry.Timestamp.Format("15:04:05")
		builder.WriteString(fmt.Sprintf("- [%s] %s\n", timestamp, entry.Command))
	}
	builder.WriteString("\nOutput exactly 2-3 bullet points capturing the likely goals, each ending with the required supporting commands in parentheses.")
	return builder.String()
}

func buildOptimizationPrompt(summary string) string {
	var builder strings.Builder
	builder.WriteString("You previously summarized the user's likely goals from today's shell commands as follows:\n")
	builder.WriteString(summary)
	builder.WriteString("\n\nBased on those goals, propose faster or more effective ways the user could accomplish them. Provide 3-5 practical, actionable suggestions. Each bullet should recommend a concrete improvement and may cover any relevant workflow enhancements. Avoid restating the original commands; focus on forward-looking recommendations.")
	return builder.String()
}

func buildWebSearchPrompt(summary, suggestions string) string {
	var builder strings.Builder
	builder.WriteString("The user's goals have been summarized as:\n")
	builder.WriteString(summary)
	if suggestions != "" {
		builder.WriteString("\n\nPrevious recommendations already provided:\n")
		builder.WriteString(suggestions)
	}
	builder.WriteString("\n\nUse web search to uncover interesting, unique, or unconventional ways to achieve these goals. Focus on additions or alternatives beyond the earlier recommendations. Return 2-3 bullet points, each highlighting what makes the approach distinctive and include the source domain in parentheses (e.g., source: example.com).")
	return builder.String()
}

func withWebSearchTool() func(*responses.ResponseNewParams) {
	return func(params *responses.ResponseNewParams) {
		params.Tools = append(params.Tools, responses.ToolParamOfWebSearchPreview(responses.WebSearchToolTypeWebSearchPreview))
		choice := responses.ToolChoiceTypesParam{
			Type: responses.ToolChoiceTypesTypeWebSearchPreview,
		}
		params.ToolChoice = responses.ResponseNewParamsToolChoiceUnion{OfHostedTool: &choice}
	}
}
