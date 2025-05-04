package models

type Model string
type Token int

const (
	ModelGPT          string = "gpt"
	ModelGPTName      Model  = "gpt-4.1"
	ModelGPTMaxTokens Token  = 32768 // 32,768

	ModelGPTo          string = "gpt-o"
	ModelGPToName      Model  = "o4-mini"
	ModelGPToMaxTokens Token  = 100000 // 100,000

	ModelClaude          string = "claude"
	ModelClaudeName      Model  = "claude-3-7-sonnet-latest"
	ModelClaudeMaxTokens Token  = 8192 // 8,192
)
