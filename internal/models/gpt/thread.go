package gpt

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ThreadResponse struct {
	ID        string   `json:"id"`
	Object    string   `json:"object"`
	CreatedAt int      `json:"created_at"`
	Metadata  Metadata `json:"metadata"`
}

// createThread creates a new thread and returns the thread ID.
func createThread(apiKey string) (string, error) {
	ctx := context.Background()
	const url = "https://api.openai.com/v1/threads"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		errorContent, _ := errorPrettyPrint(respBody)
		return "", fmt.Errorf("error making a http request. Statuscode: %s. Error: %s", resp.Status, errorContent)
	}

	var threadResponse ThreadResponse
	if err := json.Unmarshal(respBody, &threadResponse); err != nil {
		return "", err
	}

	return threadResponse.ID, nil
}
