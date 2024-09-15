package models

type Model string
type Token int

const (
	ModelGPT          string = "gpt"
	ModelGPTName      Model  = "gpt-4o-2024-08-06"
	ModelGPTMaxTokens Token  = 16384

	ModelGPTo1          string = "gpt-o1"
	ModelGPTo1Name      Model  = "o1-preview-2024-09-12"
	ModelGPTo1MaxTokens Token  = 32768

	ModelGPTo1Mini          string = "gpt-o1-mini"
	ModelGPTo1MiniName      Model  = "o1-mini-2024-09-12"
	ModelGPTo1MiniMaxTokens Token  = 65536

	ModelClaude          string = "claude"
	ModelClaudeName      Model  = "claude-3-5-sonnet-20240620"
	ModelClaudeMaxTokens Token  = 4092
)
