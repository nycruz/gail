package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"log/slog"

	"github.com/nycruz/gail/internal/models"
	"github.com/nycruz/gail/internal/validator"
)

// Claude implements the LLM interface
type Claude struct {
	// ID of the Anthropic Claude model to use for the chat completion (e.g. claude-3-opus-20240229).
	Model models.Model
	// The maximum number of tokens to generate in the chat completion.
	MaxTokens models.Token
	// Stores the "user" and "assistant" messages.
	messages []Message
	// The current persona used for the chat completion.
	currentRolePersona string
	// The current instruction used for the chat completion.
	currentSkillInstruction string
	// The Claude API Key
	apiKey string
	// Has no meaning or use. Done to satisfy interface implementation.
	user      string
	validator *validator.Validator
	Logger    *slog.Logger
}

type MessageRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type MessageResponse struct {
	ID      string `json:"id"`
	Content []struct {
		Text  string `json:"text,omitempty"`
		ID    string `json:"id,omitempty"`
		Name  string `json:"name,omitempty"`
		Input struct {
		} `json:"input,omitempty"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func New(logger *slog.Logger, apiKey string, model models.Model, maxTokens models.Token, user string, validator *validator.Validator) (*Claude, error) {
	claude := &Claude{
		Model:                   model,
		apiKey:                  apiKey,
		user:                    user,
		MaxTokens:               maxTokens,
		Logger:                  logger,
		messages:                []Message{},
		currentRolePersona:      "",
		currentSkillInstruction: "",
		validator:               validator,
	}

	return claude, nil
}

func (c *Claude) Prompt(ctx context.Context, roleName string, rolePersona string, skillInstruction string, message string) (string, error) {
	validationMsg, isValid := c.validator.Validate(message)
	if !isValid {
		return validationMsg, nil
	}

	if rolePersona != c.currentRolePersona || skillInstruction != c.currentSkillInstruction {
		c.currentRolePersona = rolePersona
		c.currentSkillInstruction = skillInstruction

		c.messages = append(c.messages, Message{
			Role:    "user",
			Content: fmt.Sprintf("%s. %s. %s", c.currentRolePersona, c.currentSkillInstruction, message),
		})
	} else {
		c.messages = append(c.messages, Message{
			Role:    "user",
			Content: message,
		})
	}

	messageRequest := MessageRequest{
		Model:     string(c.Model),
		MaxTokens: int(c.MaxTokens),
		Messages:  c.messages,
	}

	reqBody, err := json.Marshal(messageRequest)
	if err != nil {
		return "", fmt.Errorf("unable to json marshal Claude Message request: %w", err)
	}

	url := "https://api.anthropic.com/v1/messages"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("unable to create Claude Message http request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("unable to make Claude Message http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Claude Message http request failed with unexpected statuscode: %s - %s", resp.Status, string(bodyBytes))
	}

	var msr MessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&msr); err != nil {
		return "", fmt.Errorf("unable to json decode Claude Response http response: %w", err)
	}

	c.messages = append(c.messages, Message{
		Role:    "assistant",
		Content: msr.Content[0].Text,
	})

	response := msr.Content[0].Text
	return response, nil
}

// GetModel returns the model used for the chat completion.
func (c *Claude) GetModel() string {
	return string(c.Model)
}

// GetUser returns the user used for the chat completion.
func (c *Claude) GetUser() string {
	return c.user
}
