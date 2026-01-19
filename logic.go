package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ExchangeRatesResponse represents the response from exchangeratesapi.io
type ExchangeRatesResponse struct {
	Success   bool              `json:"success"`
	Timestamp int64             `json:"timestamp"`
	Base      string            `json:"base"`
	Date      string            `json:"date"`
	Rates     map[string]float64 `json:"rates"`
}

// ConversionResponse represents the response to send to the user
type ConversionResponse struct {
	From      string  `json:"from"`
	To        string  `json:"to"`
	Rate      float64 `json:"rate"`
	Date      string  `json:"date"`
	Timestamp int64   `json:"timestamp"`
}

// ConvertEURToCOP fetches exchange rates and extracts COP rate
func ConvertEURToCOP(apiKey string) (*ConversionResponse, error) {
	// Build the API URL
	url := fmt.Sprintf("https://api.exchangeratesapi.io/v1/latest?access_key=%s&base=EUR", apiKey)

	// Make HTTP request
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call exchange rates API: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("exchange rates API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse JSON response
	var exchangeResp ExchangeRatesResponse
	if err := json.Unmarshal(body, &exchangeResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Check if API call was successful
	if !exchangeResp.Success {
		return nil, fmt.Errorf("exchange rates API returned success=false")
	}

	// Extract COP rate
	copRate, exists := exchangeResp.Rates["COP"]
	if !exists {
		return nil, fmt.Errorf("COP rate not found in API response")
	}

	// Build response
	result := &ConversionResponse{
		From:      "EUR",
		To:        "COP",
		Rate:      copRate,
		Date:      exchangeResp.Date,
		Timestamp: exchangeResp.Timestamp,
	}

	return result, nil
}

