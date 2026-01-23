package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	mysql "github.com/go-sql-driver/mysql"
	"github.com/joy-currency-conversion-GCP/config"
)

var ErrEmailAlreadyExists = errors.New("favorite with this email already exists")

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

var mysqlDB *sql.DB

func InitMySQLFromEnv(cfg *config.Config) error {
	db, err := cfg.DBConfig.Connect()
	if err != nil {
		return err
	}

	mysqlDB = db

	if err := config.InitSchema(mysqlDB); err != nil {
		return err
	}

	return nil
}

func SaveFavoriteConversion(email, currencyOrigin, currencyDestination string, threshold float64) (int64, error) {
	if mysqlDB == nil {
		return 0, fmt.Errorf("database is not initialized")
	}

	res, err := mysqlDB.Exec(
		`INSERT INTO favorite_conversions (email, currency_origin, currency_destination, threshold) VALUES (?, ?, ?, ?)`,
		email, currencyOrigin, currencyDestination, threshold,
	)
	if err != nil {
		if me, ok := err.(*mysql.MySQLError); ok && me.Number == 1062 {
            return 0, ErrEmailAlreadyExists
        }
		return 0, fmt.Errorf("failed to insert favorite conversion: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get inserted id: %w", err)
	}
	return id, nil
}