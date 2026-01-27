package config

import (
	"context"
	"fmt"
	"log"
	"os"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/joho/godotenv"
)

type Environment string

const (
	EnvLocal      Environment = "local"
	EnvProduction Environment = "production"
)

type Config struct {
	Port            string
	Environment     Environment
	APIKey          string
	FunctionURL     string // URL de la Cloud Function para enviar emails
	PubSubTopicID   string // Topic ID de Pub/Sub para notificaciones
	DBConfig        DatabaseConfig
}

type DatabaseConfig struct {
	Host           string
	Port           string
	Name           string
	User           string
	Password       string
	ConnectionName string // Para Cloud SQL
}

func Load() *Config {
	env := detectEnvironment()

	log.Printf("🚀 Worker iniciando en modo: %s", env)

	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: env,
		DBConfig:    loadDatabaseConfig(env),
		APIKey:      loadSecret("EXCHANGE_RATES_API_KEY", "EXCHANGE_RATES_API_KEY", env),
		FunctionURL: loadSecret("FUNCTION_URL", "FUNCTION_URL", env),
		PubSubTopicID: loadSecret("PUBSUB_TOPIC_ID", "PUBSUB_TOPIC_ID", env),
	}

	return cfg
}

func detectEnvironment() Environment {
	if os.Getenv("GCP_PROJECT_ID") != "" {
		return EnvProduction
	}
	if os.Getenv("ENVIRONMENT") == "production" {
		return EnvProduction
	}
	return EnvLocal
}

func loadDatabaseConfig(env Environment) DatabaseConfig {
	if env == EnvProduction {
		log.Println("📦 Configurando Cloud SQL (producción)")
		return DatabaseConfig{
			ConnectionName: getEnv("CLOUD_SQL_CONNECTION_NAME", ""),
			Name:           getEnv("DB_NAME", "currency_conversion"),
			User:           loadSecret("DB_USER", "DB_USER", env),
			Password:       loadSecret("DB_PASSWORD", "DB_PASSWORD", env),
		}
	}

	log.Println("💻 Configurando MySQL local (desarrollo)")
	return DatabaseConfig{
		Host:     getEnv("DB_HOST", "mysql"),
		Port:     getEnv("DB_PORT", "3306"),
		Name:     getEnv("DB_NAME", "currency_conversion"),
		User:     getEnv("DB_USER", "app"),
		Password: getEnv("DB_PASSWORD", "app_password"),
	}
}

func loadSecret(envKey, secretName string, env Environment) string {
	if env == EnvProduction {
		log.Printf("🔐 Obteniendo '%s' desde Secret Manager", secretName)
		return getSecretFromGCP(secretName)
	}

	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, usando variables de entorno del sistema")
	}
	value := os.Getenv(envKey)
	if value == "" {
		log.Fatalf("❌ Variable %s no encontrada en ambiente local", envKey)
	}
	return value
}

func getSecretFromGCP(secretName string) string {
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		log.Fatal("❌ GCP_PROJECT_ID no configurado")
	}

	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Fatalf("❌ Error creando cliente Secret Manager: %v", err)
	}
	defer client.Close()

	name := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, secretName)
	req := &secretmanagerpb.AccessSecretVersionRequest{Name: name}

	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		log.Fatalf("❌ Error accediendo al secreto %s: %v", secretName, err)
	}

	log.Printf("✅ Secreto '%s' obtenido correctamente", secretName)
	return string(result.Payload.Data)
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func (c *Config) IsProduction() bool {
	return c.Environment == EnvProduction
}