package tui

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	tea "github.com/charmbracelet/bubbletea"
)

type Answer struct {
	msg    string
	Answer string // Answer from Gail
	Error  errMsg // Error from Gail
}

func (m model) fetchAnswer(roleName string, rolePersona string, skillInstruction string, message string) tea.Cmd {
	ctx := context.Background()

	return func() tea.Msg {
		answer, err := m.llm.Prompt(ctx, roleName, rolePersona, skillInstruction, message)
		m.logger.Info(fmt.Sprintf("LLM Answer: %v", answer))
		if err != nil {
			e := fmt.Errorf("%s: %w", m.llm.GetModel(), err)
			return Answer{Error: e}
		}

		highlightedAnswer, err := highlightCodeSnippetsAndAssembleResponse(answer)
		if err != nil {
			return Answer{Error: err}
		}

		return Answer{Answer: highlightedAnswer, msg: "Answer fetched successfully"}
	}
}

// highlightCodeSnippetsAndAssembleResponse highlights all code snippets in the response
func highlightCodeSnippetsAndAssembleResponse(response string) (string, error) {
	snippets := extractCodeSnippets(response)
	var highlightedResponse strings.Builder

	languageHint := extractProgrammingLanguageFromResponse(response)
	for i, snippet := range snippets {
		// Apply syntax highlighting to every second element (code snippets)
		if i%2 == 1 {
			highlighted, err := highlightCodeSnippetsWithChroma(snippet, languageHint)
			if err != nil {
				return "", err
			}
			highlightedResponse.WriteString(highlighted)
		} else {
			highlightedResponse.WriteString(snippet)
		}
	}

	return highlightedResponse.String(), nil
}

// highlightCodeSnippetsWithChroma applies syntax highlighting to a code snippet using Chroma.
func highlightCodeSnippetsWithChroma(codeSnippet string, languageHint string) (string, error) {
	var lexer chroma.Lexer

	if languageHint != "" {
		lexer = lexers.Get(languageHint)
	} else {
		lexer = lexers.Analyse(codeSnippet)
	}

	if lexer == nil {
		lexer = lexers.Fallback
	}

	lexer = chroma.Coalesce(lexer)

	formatter := formatters.Get("terminal256")
	if formatter == nil {
		return "", fmt.Errorf("Chroma: no formatter found")
	}

	// style gallery: https://xyproto.github.io/splash/docs/all.html
	style := styles.Get("friendly")
	if style == nil {
		return "", fmt.Errorf("Chroma: no style found")
	}

	iterator, err := lexer.Tokenise(nil, codeSnippet)
	if err != nil {
		return "", fmt.Errorf("Chroma: tokenise error: %w", err)
	}

	var b bytes.Buffer
	err = formatter.Format(&b, style, iterator)
	if err != nil {
		return "", fmt.Errorf("Chroma: format error: %w", err)
	}

	return b.String(), nil
}

// extractCodeSnippets extracts all code snippets from the response
// and returns a slice where normal text and code snippets alternate.
func extractCodeSnippets(response string) []string {
	var snippets []string
	start := 0

	for {
		startIndex := strings.Index(response[start:], "```")
		if startIndex == -1 {
			// No more code snippets, add the rest of the response
			snippets = append(snippets, response[start:])
			break
		}
		startIndex += start

		// Add text before the code snippet
		snippets = append(snippets, response[start:startIndex])

		// Find the end of the code snippet
		endIndex := strings.Index(response[startIndex+3:], "```")
		if endIndex == -1 {
			// No closing backticks found, add the rest of the response as normal text
			snippets = append(snippets, response[startIndex:])
			break
		}
		endIndex += startIndex + 3

		// Add the code snippet
		snippets = append(snippets, response[startIndex+3:endIndex])

		// Move the start position past this code snippet
		start = endIndex + 3
	}

	return snippets
}

// extractProgrammingLanguageFromResponse extracts the programming language from the response.
// It looks for a pattern like "```language" and returns the language name.
func extractProgrammingLanguageFromResponse(response string) string {
	// Regex to find patterns like "```language"
	langRegex := regexp.MustCompile("```([a-zA-Z]+)")

	matches := langRegex.FindStringSubmatch(response)
	if len(matches) >= 2 {
		// The second element in matches will be the captured group, which is the language
		return matches[1]
	}

	return ""
}
