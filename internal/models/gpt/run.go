package gpt

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type RunRequest struct {
	AssistantID string `json:"assistant_id"`
}

type RunResponse struct {
	ID           string `json:"id"`
	Object       string `json:"object"`
	CreatedAt    int    `json:"created_at"`
	AssistantID  string `json:"assistant_id"`
	ThreadID     string `json:"thread_id"`
	Status       string `json:"status"`
	StartedAt    int    `json:"started_at"`
	ExpiresAt    any    `json:"expires_at"`
	CancelledAt  any    `json:"cancelled_at"`
	FailedAt     any    `json:"failed_at"`
	CompletedAt  int    `json:"completed_at"`
	LastError    any    `json:"last_error"`
	Model        string `json:"model"`
	Instructions any    `json:"instructions"`
	Tools        []struct {
		Type string `json:"type"`
	} `json:"tools"`
	FileIds  []string `json:"file_ids"`
	Metadata struct {
	} `json:"metadata"`
}

func (gpt *GPT) createRun(ctx context.Context) (string, error) {
	runRequest := RunRequest{
		AssistantID: gpt.AssistantID,
	}

	reqBody, err := json.Marshal(runRequest)
	if err != nil {
		return "", fmt.Errorf("unable to json marshal the request: %w", err)
	}

	url := fmt.Sprintf("https://api.openai.com/v1/threads/%s/runs", gpt.ThreadID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
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

	var rr RunResponse
	if err := json.NewDecoder(resp.Body).Decode(&rr); err != nil {
		return "", fmt.Errorf("unable to decode the response body: %w", err)
	}

	return rr.ID, nil
}

func (gpt *GPT) waitRunCompleted(ctx context.Context, runID string) error {
	retryLimit := 20
	waitTime := 4 * time.Second
	retryCount := 0
	isCompleted := false
	completedStatus := "completed"

	for !isCompleted && retryCount < retryLimit {
		url := fmt.Sprintf("https://api.openai.com/v1/threads/%s/runs/%s", gpt.ThreadID, runID)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

		var rr RunResponse
		if err := json.NewDecoder(resp.Body).Decode(&rr); err != nil {
			return fmt.Errorf("unable to decode the response body: %w", err)
		}

		gpt.Logger.Info("GPT: WaitRunCompleted", "run_num", retryCount, "status", rr.Status)

		if rr.Status == completedStatus {
			return nil
		}

		time.Sleep(waitTime)
		retryCount++
	}

	return fmt.Errorf("Run did not complete with status '%s' after %d retries every %d seconds.", completedStatus, retryLimit, waitTime)
}
