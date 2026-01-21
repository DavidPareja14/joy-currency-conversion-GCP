package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	mysql "github.com/go-sql-driver/mysql"
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

func InitMySQLFromEnv() error {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "mysql"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "3306"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "app"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "app_password"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "currency_conversion"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true", user, password, host, port, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("sql.Open failed: %w", err)
	}

	// Simple retry loop because MySQL container might not be ready immediately
	var pingErr error
	for i := 0; i < 30; i++ {
		pingErr = db.Ping()
		if pingErr == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if pingErr != nil {
		_ = db.Close()
		return fmt.Errorf("db ping failed after retries: %w", pingErr)
	}

	mysqlDB = db

	// Create table if it doesn't exist
	_, err = mysqlDB.Exec(`
CREATE TABLE IF NOT EXISTS favorite_conversions (
  id BIGINT NOT NULL AUTO_INCREMENT,
  email VARCHAR(255) NOT NULL,
  currency_origin VARCHAR(10) NOT NULL,
  currency_destination VARCHAR(10) NOT NULL,
  threshold DOUBLE NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY unique_email (email)
);`)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
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