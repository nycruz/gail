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

type MessageRequest struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type MessageResponse struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	CreatedAt int    `json:"created_at"`
	ThreadID  string `json:"thread_id"`
	Role      string `json:"role"`
	Content   []struct {
		Type string `json:"type"`
		Text struct {
			Value       string `json:"value"`
			Annotations []any  `json:"annotations"`
		} `json:"text"`
	} `json:"content"`
	FileIds     []string `json:"file_ids"`
	AssistantID string   `json:"assistant_id"`
	RunID       string   `json:"run_id"`
	Metadata    struct {
	} `json:"metadata"`
}

type MessagesResponse struct {
	Object string `json:"object"`
	Data   []struct {
		ID        string `json:"id"`
		Object    string `json:"object"`
		CreatedAt int    `json:"created_at"`
		ThreadID  string `json:"thread_id"`
		Role      string `json:"role"`
		Content   []struct {
			Type string `json:"type"`
			Text struct {
				Value       string `json:"value"`
				Annotations []any  `json:"annotations"`
			} `json:"text"`
		} `json:"content"`
		FileIds     []any `json:"file_ids"`
		AssistantID any   `json:"assistant_id"`
		RunID       any   `json:"run_id"`
		Metadata    struct {
		} `json:"metadata"`
	} `json:"data"`
	FirstID string `json:"first_id"`
	LastID  string `json:"last_id"`
	HasMore bool   `json:"has_more"`
}

// createMessage creates a new OpenAI Message with the given content.
func (gpt *GPT) createMessage(ctx context.Context, message string) error {
	messageRequest := MessageRequest{
		Role:    "user",
		Content: message,
	}

	reqBody, err := json.Marshal(messageRequest)
	if err != nil {
		return fmt.Errorf("unable to json marshal the request: %w", err)
	}

	url := fmt.Sprintf("https://api.openai.com/v1/threads/%s/messages", gpt.ThreadID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("unable to create the http request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+gpt.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to make the http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.New("unable to read the http response body")
		}
		return fmt.Errorf("the http request failed with unexpected statuscode: %s. %s", resp.Status, string(bodyBytes))
	}

	return nil
}

// getResponse retrieves the response from the OpenAI API.
func (gpt *GPT) getResponse(ctx context.Context) (string, error) {
	url := fmt.Sprintf("https://api.openai.com/v1/threads/%s/messages", gpt.ThreadID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

	var msr MessagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&msr); err != nil {
		return "", fmt.Errorf("unable to decode message response: %w", err)
	}

	response := msr.Data[0].Content[0].Text.Value

	return response, nil
}
