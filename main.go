package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

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

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Get API key from environment
	apiKey := os.Getenv("EXCHANGE_RATES_API_KEY")
	if apiKey == "" {
		log.Println("EXCHANGE_RATES_API_KEY environment variable is required")
	}

	http.HandleFunc("/convert", convertHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func convertHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("EXCHANGE_RATES_API_KEY")

	if apiKey == "" {
		projectId := os.Getenv("PROJECT_ID")
		apiKey = getSecret("projects/" + projectId + "/secrets/EXCHANGE_RATES_API_KEY/versions/latest")
	}
	// Call the conversion logic
	result, err := ConvertEURToCOP(apiKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set content type and return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
