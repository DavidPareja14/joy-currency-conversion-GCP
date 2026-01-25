package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type HTTPNotifier struct {
	functionURL string
}

func NewHTTPNotifier(functionURL string) *HTTPNotifier {
	return &HTTPNotifier{
		functionURL: functionURL,
	}
}

func (n *HTTPNotifier) SendNotification(ctx context.Context, notification EmailNotification) error {
	body, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("error marshaling notification: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.functionURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error calling cloud function: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("cloud function returned status %d", resp.StatusCode)
	}

	return nil
}