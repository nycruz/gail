package models

type Model string
type Token int

const (
	ModelGPT          string = "gpt"
	ModelGPTName      Model  = "gpt-4o"
	ModelGPTMaxTokens Token  = 16384

	ModelGPTo1          string = "gpt-o1"
	ModelGPTo1Name      Model  = "o1"
	ModelGPTo1MaxTokens Token  = 100000

	ModelGPTo1Mini          string = "gpt-o1-mini"
	ModelGPTo1MiniName      Model  = "o1-mini"
	ModelGPTo1MiniMaxTokens Token  = 65536

	ModelClaude          string = "claude"
	ModelClaudeName      Model  = "claude-3-5-sonnet-latest"
	ModelClaudeMaxTokens Token  = 8192
)
