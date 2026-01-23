package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	
	"github.com/joy-currency-conversion-GCP/config"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

func getSecret(secretName string) string {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretName, // projects/PROJECT_ID/secrets/API_KEY/versions/latest
	}

	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		log.Fatalf("failed to access secret: %v", err)
	}

	return string(result.Payload.Data)
}

var appConfig *config.Config

func main() {
	appConfig = config.Load()
	if err := InitMySQLFromEnv(appConfig); err != nil {
		log.Fatalf("failed to initialize MySQL: %v", err)
	}

	http.HandleFunc("/convert", convertHandler)
	http.HandleFunc("/favorites", favoritesHandler)

	addr := ":" + appConfig.Port
	log.Printf("🚀 Servidor escuchando en %s (ambiente: %s)", addr, appConfig.Environment)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("❌ Error iniciando servidor: %v", err)
	}
}

func convertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, err := ConvertEURToCOP(appConfig.APIKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set content type and return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

type FavoriteRequest struct {
	Email               string  `json:"email"`
	CurrencyOrigin      string  `json:"currency_origin"`
	CurrencyDestination string  `json:"currency_destination"`
	Threshold           float64 `json:"threshold"`
}

type FavoriteResponse struct {
	ID                  int64   `json:"id"`
	Email               string  `json:"email"`
	CurrencyOrigin      string  `json:"currency_origin"`
	CurrencyDestination string  `json:"currency_destination"`
	Threshold           float64 `json:"threshold"`
}

func favoritesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req FavoriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.CurrencyDestination == "" {
		http.Error(w, "email and currency_destination are required", http.StatusBadRequest)
		return
	}

	if req.CurrencyOrigin == "" {
		req.CurrencyOrigin = "EUR"
	}
	if req.CurrencyOrigin != "EUR" {
		http.Error(w, "currency_origin must be EUR", http.StatusBadRequest)
		return
	}

	id, err := SaveFavoriteConversion(req.Email, req.CurrencyOrigin, req.CurrencyDestination, req.Threshold)
	if err != nil {
		if err == ErrEmailAlreadyExists {
			http.Error(w, "favorite with this email already exists", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(FavoriteResponse{
		ID:                  id,
		Email:               req.Email,
		CurrencyOrigin:      req.CurrencyOrigin,
		CurrencyDestination: req.CurrencyDestination,
		Threshold:           req.Threshold,
	})
}
