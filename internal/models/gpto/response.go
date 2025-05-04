package gpto

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type ResponseRequest struct {
	Model           string `json:"model"`
	Instructions    string `json:"instructions"`
	Input           string `json:"input"`
	User            string `json:"user"`
	MaxOutputTokens int    `json:"max_output_tokens"`
	Reasoning       struct {
		Effort string `json:"effort"` // Effort can be "low", "medium", or "high"
	} `json:"reasoning"`
}

type ResponseResponse struct {
	Status            string `json:"status"`
	Error             any    `json:"error"`
	IncompleteDetails any    `json:"incomplete_details"`
	Output            []struct {
		Type    string `json:"type"`
		ID      string `json:"id"`
		Status  string `json:"status"`
		Role    string `json:"role"`
		Content []struct {
			Type        string `json:"type"`
			Text        string `json:"text"`
			Annotations []any  `json:"annotations"`
		} `json:"content"`
	} `json:"output"`
}

func (gpto *GPTO) response(ctx context.Context, persona string, instruction string, message string, effort string) (string, error) {
	responseRequest := ResponseRequest{
		Model:           string(gpto.Model),
		Instructions:    fmt.Sprintf("%s. %s.", persona, instruction),
		Input:           message,
		User:            gpto.User,
		MaxOutputTokens: int(gpto.MaxTokens),
		Reasoning: struct {
			Effort string `json:"effort"`
		}{
			Effort: effort,
		},
	}

	reqBody, err := json.Marshal(responseRequest)
	if err != nil {
		return "", fmt.Errorf("unable to json marshal the request: %w", err)
	}

	const url = "https://api.openai.com/v1/responses"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("unable to create the http request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+gpto.apiKey)
	req.Header.Set("Content-Type", "application/json")

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

	var responseResponse ResponseResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseResponse); err != nil {
		return "", fmt.Errorf("unable to json decode the response body: %w", err)
	}

	response, err := validateResponse(&responseResponse)
	if err != nil {
		return "", fmt.Errorf("unable to get the answer from the response: %w", err)
	}

	return response, nil
}

// validateResponse checks if the response is valid and returns the answer as expected
func validateResponse(r *ResponseResponse) (string, error) {
	if r == nil {
		return "", fmt.Errorf("nil response")
	}
	if len(r.Output) <= 1 {
		return "", fmt.Errorf("expected at least 2 Outputs from OpenAI, got %d", len(r.Output))
	}
	contents := r.Output[1].Content
	if len(contents) == 0 {
		return "", fmt.Errorf("no content in Output[1]")
	}
	if r.Status != "completed" {
		return "", fmt.Errorf("OpenAI's response is not completed: status: %s, error: %s, details: %s", r.Status, r.Error, r.IncompleteDetails)
	}

	return contents[0].Text, nil
}
