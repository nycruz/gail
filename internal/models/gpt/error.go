package gpt

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// errorPrettyPrint pretty prints the error response from the OpenAI' API.
func errorPrettyPrint(body []byte) (string, error) {
	content := ""

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "\t"); err != nil {
		return "", fmt.Errorf("could not create error pretty printing: %w", err)
	}

	content = prettyJSON.String()

	return content, nil
}
