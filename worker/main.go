package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/joy-currency-conversion-GCP/worker/config"
)

var appConfig *config.Config

func main() {
	appConfig = config.Load()

	if err := InitDB(appConfig); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	InitNotifier(appConfig)

	http.HandleFunc("/check-thresholds", checkThresholdsHandler)
	http.HandleFunc("/health", healthHandler)

	addr := ":" + appConfig.Port
	log.Printf("🚀 Worker running on %s (environment: %s)", addr, appConfig.Environment)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("❌ Error starting server: %v", err)
	}
}

type CheckResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

func checkThresholdsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("📊 Starting threshold check...")

	if err := CheckThresholdsAndNotify(r.Context(), appConfig.APIKey); err != nil {
		log.Printf("Error checking thresholds: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(CheckResponse{
			Message: "error checking thresholds",
			Status:  "error",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CheckResponse{
		Message: "thresholds checked successfully",
		Status:  "success",
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}