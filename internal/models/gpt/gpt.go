package gpt

import (
	"context"
	"errors"
	"fmt"

	"log/slog"

	"github.com/nycruz/gail/internal/models"
	"github.com/nycruz/gail/internal/validator"
)

// GPT implements the LLM interface
type GPT struct {
	// ID of the OpenAI model to use for the chat completion (e.g. gpt-3.5-turbo).
	Model models.Model
	// A unique identifier representing your end-user, which can help OpenAI to monitor and detect abuse.
	User string
	// The maximum number of tokens to generate in the chat completion.
	// A rule of thumb is that one token generally corresponds to ~4 characters of text.
	MaxTokens models.Token
	// OpenAI thread ID.
	ThreadID string
	// OpenAI assistant ID.
	AssistantID string
	// The OpenAI API key.
	apiKey string
	// The current persona used for the chat completion.
	currentRolePersona string
	// The current instruction used for the chat completion.
	currentSkillInstruction string
	// The validator used to validate the input message.
	validator *validator.Validator
	// The logger used for logging messages.
	Logger *slog.Logger
}

func New(logger *slog.Logger, apiKey string, model models.Model, maxTokens models.Token, user string, validator *validator.Validator) (*GPT, error) {
	threadID, err := createThread(apiKey)
	if err != nil {
		return nil, fmt.Errorf("could not to create an OpenAI Thread: %w", err)
	}

	gpt := &GPT{
		Model:                   model,
		User:                    user,
		apiKey:                  apiKey,
		MaxTokens:               maxTokens,
		ThreadID:                threadID,
		currentRolePersona:      "",
		currentSkillInstruction: "",
		validator:               validator,
		Logger:                  logger,
	}

	return gpt, nil
}

func (gpt *GPT) Prompt(ctx context.Context, roleName string, rolePersona string, skillInstruction string, message string) (string, error) {
	validationMsg, isValid := gpt.validator.Validate(message)
	if !isValid {
		return validationMsg, nil
	}

	if gpt.ThreadID == "" {
		return "", errors.New("OpenAI's Thread ID is empty. No Thread has been created")
	}

	if rolePersona != gpt.currentRolePersona || skillInstruction != gpt.currentSkillInstruction {
		assistantID, err := gpt.createAssistant(ctx, roleName, rolePersona, skillInstruction)
		if err != nil {
			return "", fmt.Errorf("failed to create an OpenAI Assistant: %w", err)
		}

		gpt.AssistantID = assistantID
		gpt.currentRolePersona = rolePersona
		gpt.currentSkillInstruction = skillInstruction
	}

	if gpt.AssistantID == "" {
		return "", errors.New("OpenAI's Assistant ID is empty. No Assistant has been created")
	}

	if err := gpt.createMessage(ctx, message); err != nil {
		return "", fmt.Errorf("failed to create an OpenAI Message: %w", err)
	}

	runID, err := gpt.createRun(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create an OpenAI Run: %w", err)
	}

	if err := gpt.waitRunCompleted(ctx, runID); err != nil {
		return "", fmt.Errorf("failed to poll an OpenAI Run: %w", err)
	}

	response, err := gpt.getResponse(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get OpenAI's response: %w", err)
	}

	return response, nil

}

// GetModel returns the model used for the chat completion.
func (gpt *GPT) GetModel() string {
	return string(gpt.Model)
}

// GetUser returns the user used for the chat completion.
func (gpt *GPT) GetUser() string {
	return gpt.User
}
