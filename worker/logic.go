package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/joy-currency-conversion-GCP/worker/config"
)

var db *sql.DB
var notifier Notifier

// Notifier defines the contract for sending notifications
type Notifier interface {
	SendNotification(ctx context.Context, notification EmailNotification) error
}

// EmailNotification represents the data for sending an email
type EmailNotification struct {
	Email               string  `json:"email"`
	CurrencyOrigin      string  `json:"currency_origin"`
	CurrencyDestination string  `json:"currency_destination"`
	CurrentRate         float64 `json:"current_rate"`
	Threshold           float64 `json:"threshold"`
}

func InitDB(cfg *config.Config) error {
	conn, err := cfg.DBConfig.Connect()
	if err != nil {
		return err
	}
	db = conn
	return nil
}

func InitNotifier(cfg *config.Config) {
	notifier = NewHTTPNotifier(cfg.FunctionURL)
}

type FavoriteConversion struct {
	ID                  int64
	Email               string
	CurrencyOrigin      string
	CurrencyDestination string
	Threshold           float64
}

type ExchangeRatesResponse struct {
	Success   bool               `json:"success"`
	Timestamp int64              `json:"timestamp"`
	Base      string             `json:"base"`
	Date      string             `json:"date"`
	Rates     map[string]float64 `json:"rates"`
}

func GetAllFavorites() ([]FavoriteConversion, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := db.Query(`
		SELECT id, email, currency_origin, currency_destination, threshold 
		FROM favorite_conversions
	`)
	if err != nil {
		return nil, fmt.Errorf("error querying favorites: %w", err)
	}
	defer rows.Close()

	var favorites []FavoriteConversion
	for rows.Next() {
		var fav FavoriteConversion
		err := rows.Scan(&fav.ID, &fav.Email, &fav.CurrencyOrigin, &fav.CurrencyDestination, &fav.Threshold)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		favorites = append(favorites, fav)
	}

	return favorites, nil
}

func GetExchangeRate(apiKey, base, target string) (float64, error) {
	url := fmt.Sprintf("https://api.exchangeratesapi.io/v1/latest?access_key=%s&base=%s", apiKey, base)

	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to call exchange rates API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("exchange rates API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	var exchangeResp ExchangeRatesResponse
	if err := json.Unmarshal(body, &exchangeResp); err != nil {
		return 0, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	if !exchangeResp.Success {
		return 0, fmt.Errorf("exchange rates API returned success=false")
	}

	rate, exists := exchangeResp.Rates[target]
	if !exists {
		return 0, fmt.Errorf("%s rate not found in API response", target)
	}

	return rate, nil
}

func CheckThresholdsAndNotify(ctx context.Context, apiKey string) error {
	favorites, err := GetAllFavorites()
	if err != nil {
		return fmt.Errorf("error getting favorites: %w", err)
	}

	log.Printf("Found %d favorite conversions to check", len(favorites))

	maxChecks := 1 // For avoid rate limit in the free tier
	if len(favorites) > maxChecks {
		log.Printf("⚠️ Limiting checks to %d favorites (found %d total)", maxChecks, len(favorites))
		favorites = favorites[:maxChecks]
	}

	for _, fav := range favorites {
		rate, err := GetExchangeRate(apiKey, fav.CurrencyOrigin, fav.CurrencyDestination)
		if err != nil {
			log.Printf("Error getting rate for %s->%s (email: %s): %v", 
				fav.CurrencyOrigin, fav.CurrencyDestination, fav.Email, err)
			continue
		}

		log.Printf("Checking %s: rate=%.2f, threshold=%.2f", fav.Email, rate, fav.Threshold)

		if rate >= fav.Threshold {
			notification := EmailNotification{
				Email:               fav.Email,
				CurrencyOrigin:      fav.CurrencyOrigin,
				CurrencyDestination: fav.CurrencyDestination,
				CurrentRate:         rate,
				Threshold:           fav.Threshold,
			}

			if err := notifier.SendNotification(ctx, notification); err != nil {
				log.Printf("Error sending notification to %s: %v", fav.Email, err)
				continue
			}

			log.Printf("✅ Notification sent to %s (rate %.2f >= threshold %.2f)", 
				fav.Email, rate, fav.Threshold)
		}
	}

	return nil
}