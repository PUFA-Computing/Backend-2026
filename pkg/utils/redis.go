package utils

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"time"
)

var Rdb *redis.Client

// RedisEnabled indicates if Redis is available and connected
var RedisEnabled bool = false

func InitRedis() {
	redisURL := os.Getenv("REDIS_URL")
	redisPassword := os.Getenv("REDIS_PASS")

	log.Printf("Redis URL: %s", redisURL)

	// Skip Redis initialization if URL is empty
	if redisURL == "" {
		log.Println("Redis URL not provided. Redis functionality will be disabled.")
		RedisEnabled = false
		return
	}

	options := &redis.Options{
		Addr:         redisURL,
		Password:     redisPassword,
		DB:           0,
		DialTimeout:  10 * time.Second,  // Connection timeout
		ReadTimeout:  10 * time.Second,  // Read timeout
		WriteTimeout: 10 * time.Second,  // Write timeout
		PoolSize:     50,                // Increased pool size for high concurrency (default is 10)
		MinIdleConns: 10,                // Minimum idle connections
		PoolTimeout:  15 * time.Second,  // Pool timeout
	}

	log.Println("Attempting to connect to Redis...")
	Rdb = redis.NewClient(options)
	
	// Implement retry mechanism
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		// Use a context with timeout for the ping
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		if err := Rdb.Ping(ctx).Err(); err != nil {
			log.Printf("Attempt %d: Failed to connect to Redis: %v", i+1, err)
			if i == maxRetries-1 {
				log.Println("All Redis connection attempts failed. Application will continue without Redis. Token revocation will not work.")
				RedisEnabled = false
				return
			}
			log.Printf("Retrying in 2 seconds...")
			time.Sleep(2 * time.Second) // Wait before retrying
			continue
		}
		
		RedisEnabled = true
		log.Println("Successfully connected to Redis")
		return
	}
}

func IsTokenRevoked(tokenString string) (bool, error) {
	if !RedisEnabled || Rdb == nil {
		// If Redis is not available, assume token is not revoked
		return false, nil
	}
	
	// Use context with timeout to avoid hanging
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	exists, err := Rdb.SIsMember(ctx, "revoked_tokens", tokenString).Result()
	if err != nil {
		log.Printf("Error checking if token is revoked: %v", err)
		// If Redis fails, assume token is not revoked to allow users to continue using the app
		return false, nil
	}

	return exists, nil
}

func RevokeToken(tokenString string) error {
	if !RedisEnabled || Rdb == nil {
		// If Redis is not available, just log and return success
		log.Println("WARNING: Redis not available, token revocation not persisted")
		return nil
	}
	
	// Use context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Try up to 3 times to revoke the token
	for i := 0; i < 3; i++ {
		_, err := Rdb.SAdd(ctx, "revoked_tokens", tokenString).Result()
		if err != nil {
			log.Printf("Attempt %d: Failed to revoke token: %v", i+1, err)
			if i < 2 {
				time.Sleep(1 * time.Second)
				continue
			}
			// After 3 attempts, log but don't fail the operation
			log.Println("WARNING: Could not persist token revocation to Redis after multiple attempts")
			return nil
		}
		break
	}

	return nil
}

func CloseRedis() {
	if !RedisEnabled || Rdb == nil {
		return
	}
	
	// Use context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Flush any pending commands
	if err := Rdb.FlushDB(ctx).Err(); err != nil {
		log.Printf("Warning: Failed to flush Redis DB: %v", err)
	}
	
	// Close the connection
	err := Rdb.Close()
	if err != nil {
		log.Printf("Error closing Redis connection: %v", err)
		return
	}
	
	log.Println("Redis connection closed successfully")
	RedisEnabled = false
	Rdb = nil
}

// Cache helper functions

// GetFromRedis retrieves a value from Redis cache
func GetFromRedis(key string) (string, error) {
	if !RedisEnabled || Rdb == nil {
		return "", redis.Nil
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	val, err := Rdb.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	
	return val, nil
}

// SetToRedis stores a value in Redis cache with TTL
func SetToRedis(key string, value string, ttl time.Duration) error {
	if !RedisEnabled || Rdb == nil {
		return nil // Silently skip if Redis not available
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	return Rdb.Set(ctx, key, value, ttl).Err()
}

// DeleteFromRedis removes a key from Redis cache
func DeleteFromRedis(key string) error {
	if !RedisEnabled || Rdb == nil {
		return nil
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	return Rdb.Del(ctx, key).Err()
}

// DeletePatternFromRedis removes all keys matching a pattern
func DeletePatternFromRedis(pattern string) error {
	if !RedisEnabled || Rdb == nil {
		return nil
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Use SCAN to find matching keys
	iter := Rdb.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := Rdb.Del(ctx, iter.Val()).Err(); err != nil {
			log.Printf("Error deleting key %s: %v", iter.Val(), err)
		}
	}
	
	return iter.Err()
}
