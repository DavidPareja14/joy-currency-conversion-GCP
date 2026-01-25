package config

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func (db *DatabaseConfig) GetDSN() string {
	if db.ConnectionName != "" {
		socketPath := fmt.Sprintf("/cloudsql/%s", db.ConnectionName)
		dsn := fmt.Sprintf("%s:%s@unix(%s)/%s?parseTime=true",
			db.User,
			db.Password,
			socketPath,
			db.Name,
		)
		log.Printf("📡 DSN Cloud SQL: %s@unix(/cloudsql/...)/%s", db.User, db.Name)
		return dsn
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
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

	conn.SetMaxOpenConns(10)
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