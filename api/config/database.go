// config/database.go
package config

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

func (db *DatabaseConfig) GetDSN() string {
	if db.ConnectionName != "" {
		socketPath := fmt.Sprintf("/cloudsql/%s", db.ConnectionName)
		dsn := fmt.Sprintf("%s:%s@unix(%s)/%s?parseTime=true&multiStatements=true",
			db.User,
			db.Password,
			socketPath,
			db.Name,
		)
		log.Printf("📡 DSN Cloud SQL: %s@unix(/cloudsql/...)/%s", db.User, db.Name)
		return dsn
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true",
		db.User,
		db.Password,
		db.Host,
		db.Port,
		db.Name,
	)
	log.Printf("📡 DSN Local: %s@tcp(%s:%s)/%s", db.User, db.Host, db.Port, db.Name)
	return dsn
}

func (db *DatabaseConfig) Connect() (*sql.DB, error) {
	dsn := db.GetDSN()

	log.Println("🔌 Conectando a MySQL...")

	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error abriendo conexión: %w", err)
	}

	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	var pingErr error
	for i := 0; i < 30; i++ {
		pingErr = conn.Ping()
		if pingErr == nil {
			break
		}
		log.Printf("⏳ Esperando MySQL... intento %d/30", i+1)
		time.Sleep(2 * time.Second)
	}

	if pingErr != nil {
		conn.Close()
		return nil, fmt.Errorf("error conectando después de reintentos: %w", pingErr)
	}

	log.Println("✅ Conexión a MySQL establecida")
	return conn, nil
}

// InitSchema crea las tablas necesarias
func InitSchema(db *sql.DB) error {
	log.Println("🔧 Inicializando esquema de base de datos...")

	_, err := db.Exec(`
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
		return fmt.Errorf("error creando tabla: %w", err)
	}

	log.Println("✅ Esquema inicializado correctamente")
	return nil
}