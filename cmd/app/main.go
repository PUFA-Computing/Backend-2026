package main

import (
	"Backend/api"
	"Backend/configs"
	"Backend/internal/database"
	"Backend/pkg/utils"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "Backend/docs" // Import generated docs
)

// @title PUFA Backend API
// @version 1.0
// @description Backend API for PUFA Computer Science with Candidates and Votes Management
// @termsOfService https://compsci.president.ac.id/terms

// @contact.name API Support
// @contact.url https://compsci.president.ac.id/support
// @contact.email support@pufacomputing.ac.id

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func tryInitRedis() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("WARNING: Redis initialization failed: %v", r)
			log.Println("Application will continue without Redis. Token revocation will not work.")
		}
	}()
	
	utils.InitRedis()
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	config := configs.LoadConfig()

	database.Migrate()
	database.Init(config)
	
	// Try to initialize Redis, but continue if it fails
	tryInitRedis()

	r := api.SetupRoutes()

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	// Start server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		// port 8080
		port := "0.0.0.0:8080"
		log.Printf("Server starting on %s", port)
		serverErr <- r.Run(port)
	}()
	
	// Wait for interrupt signal or server error
	select {
		case err := <-serverErr:
			log.Fatalf("Failed to run server: %v", err)
		case <-quit:
			log.Println("Server is shutting down...")
			
			// Close Redis connection
			log.Println("Closing Redis connection...")
			utils.CloseRedis()
			
			// Close database connection
			log.Println("Closing database connection...")
			database.Close()
			
			log.Println("Server exited properly")
	}
}
