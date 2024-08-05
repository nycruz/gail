package models

type Model string
type Token int

const (
	ModelGPT              string = "gpt"
	ModelGPTName          Model  = "gpt-4o"
	ModelGPT0125MaxTokens Token  = 4096

	ModelClaude          string = "claude"
	ModelClaudeName      Model  = "claude-3-5-sonnet-20240620"
	ModelClaudeMaxTokens Token  = 4092
)
