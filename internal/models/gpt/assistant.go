package gpt

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type AssistantRequest struct {
	Name         string   `json:"name"`
	Description  any      `json:"description"`
	Model        string   `json:"model"`
	Instructions string   `json:"instructions"`
	Tools        []Tool   `json:"tools"`
	Metadata     Metadata `json:"metadata"`
}

type AssistantResponse struct {
	ID           string   `json:"id"`
	Object       string   `json:"object"`
	CreatedAt    int      `json:"created_at"`
	Name         string   `json:"name"`
	Description  any      `json:"description"`
	Model        string   `json:"model"`
	Instructions string   `json:"instructions"`
	Tools        []Tool   `json:"tools"`
	Metadata     Metadata `json:"metadata"`
}

type Tool struct {
	Type string `json:"type"`
}

type Metadata struct{}

// createAssistant creates a new OpenAI Assistant with a given role persona and skill instruction.
func (gpt *GPT) createAssistant(ctx context.Context, roleName string, persona string, instruction string) (string, error) {
	assistantRequest := AssistantRequest{
		Name:         roleName,
		Description:  fmt.Sprintf("Gail: %s", persona),
		Model:        string(gpt.Model),
		Instructions: fmt.Sprintf("%s. %s.", persona, instruction),
		Tools: []Tool{
			{
				Type: "code_interpreter",
			},
		},
	}

	reqBody, err := json.Marshal(assistantRequest)
	if err != nil {
		return "", fmt.Errorf("unable to json marshal the request: %w", err)
	}

	const url = "https://api.openai.com/v1/assistants"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("unable to create the http request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+gpt.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("unable to make the http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", errors.New("unable to read the http response body")
		}
		return "", fmt.Errorf("the http request failed with unexpected statuscode: %s. %s", resp.Status, string(bodyBytes))
	}

	var assistantResponse AssistantResponse
	if err := json.NewDecoder(resp.Body).Decode(&assistantResponse); err != nil {
		return "", fmt.Errorf("unable to json decode the response body: %w", err)
	}

	return assistantResponse.ID, nil
}
