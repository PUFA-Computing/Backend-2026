package database

import (
	"Backend/configs"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"time"
)

type DBInterface interface {
	Query(query string, args ...interface{}) error
}

var DB *pgxpool.Pool

func Init(config *configs.Config) {
	var err error

	// Create a configuration object
	poolConfig, err := pgxpool.ParseConfig("user=" + config.DBUser +
		" password=" + config.DBPassword +
		" host=" + config.DBHost +
		" port=" + config.DBPort +
		" dbname=" + config.DBName +
		" sslmode=disable")
	
	if err != nil {
		log.Fatalf("Failed to parse database config: %v", err)
	}
	
	// Configure connection pool for high concurrency and reliability
	poolConfig.MaxConns = 100                      // Increased from 50 to handle more concurrent users
	poolConfig.MinConns = 10                       // Keep minimum connections ready
	poolConfig.MaxConnLifetime = 1 * time.Hour     // Increased from 30min for better connection reuse
	poolConfig.MaxConnIdleTime = 30 * time.Minute  // Increased from 5min to reduce reconnection overhead
	poolConfig.HealthCheckPeriod = 1 * time.Minute // Regular health checks
	
	// Create the connection pool with the enhanced configuration
	DB, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	log.Printf("Connected to database with optimized pool (MaxConns: %d, MinConns: %d)", 
		poolConfig.MaxConns, poolConfig.MinConns)
}

func Close() {
	DB.Close()
}

func GetDB() *pgxpool.Pool {
	return DB
}
