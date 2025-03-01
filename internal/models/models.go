package models

type Model string
type Token int

const (
	ModelGPT          string = "gpt"
	ModelGPTName      Model  = "gpt-4o"
	ModelGPTMaxTokens Token  = 16384

	ModelClaude          string = "claude"
	ModelClaudeName      Model  = "claude-3-7-sonnet-latest"
	ModelClaudeMaxTokens Token  = 8192
)
