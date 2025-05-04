package gpto

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nycruz/gail/internal/models"
	"github.com/nycruz/gail/internal/validator"
)

// GPTO implements the LLM interface
type GPTO struct {
	// ID of the OpenAI model to use for the chat completion (e.g. gpt-3.5-turbo).
	Model models.Model
	// A unique identifier representing your end-user, which can help OpenAI to monitor and detect abuse.
	User string
	// The maximum number of tokens to generate in the chat completion.
	// A rule of thumb is that one token generally corresponds to ~4 characters of text.
	MaxTokens models.Token
	// The OpenAI API key used for authentication.
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

func New(logger *slog.Logger, apiKey string, model models.Model, maxTokens models.Token, user string, validator *validator.Validator) (*GPTO, error) {
	gpto := &GPTO{
		Model:                   model,
		User:                    user,
		apiKey:                  apiKey,
		MaxTokens:               maxTokens,
		currentRolePersona:      "",
		currentSkillInstruction: "",
		validator:               validator,
		Logger:                  logger,
	}

	return gpto, nil
}

func (gpto *GPTO) Prompt(ctx context.Context, roleName string, rolePersona string, skillInstruction string, message string) (string, error) {
	validationMsg, isValid := gpto.validator.Validate(message)
	if !isValid {
		return validationMsg, nil
	}

	response, err := gpto.response(ctx, rolePersona, skillInstruction, message, "medium")
	if err != nil {
		return "", fmt.Errorf("failed to get OpenAI's response: %w", err)
	}

	return response, nil
}

// GetModel returns the model used for the chat completion.
func (gpto *GPTO) GetModel() string {
	return string(gpto.Model)
}

// GetUser returns the user used for the chat completion.
func (gpto *GPTO) GetUser() string {
	return gpto.User
}
