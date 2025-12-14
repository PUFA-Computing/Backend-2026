package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"Backend/pkg/utils"
)

// CacheService provides caching functionality using Redis
type CacheService struct {
	defaultTTL time.Duration
}

// NewCacheService creates a new cache service
func NewCacheService(defaultTTL time.Duration) *CacheService {
	return &CacheService{
		defaultTTL: defaultTTL,
	}
}

// Get retrieves a value from cache
func (cs *CacheService) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := utils.GetFromRedis(key)
	if err != nil {
		return fmt.Errorf("cache miss: %w", err)
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	return nil
}

// Set stores a value in cache with default TTL
func (cs *CacheService) Set(ctx context.Context, key string, value interface{}) error {
	return cs.SetWithTTL(ctx, key, value, cs.defaultTTL)
}

// SetWithTTL stores a value in cache with custom TTL
func (cs *CacheService) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	if err := utils.SetToRedis(key, string(data), ttl); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// Delete removes a value from cache
func (cs *CacheService) Delete(ctx context.Context, key string) error {
	return utils.DeleteFromRedis(key)
}

// DeletePattern removes all keys matching a pattern
func (cs *CacheService) DeletePattern(ctx context.Context, pattern string) error {
	return utils.DeletePatternFromRedis(pattern)
}

// GenerateKey creates a cache key from components
func GenerateKey(components ...string) string {
	key := "cache"
	for _, component := range components {
		key += ":" + component
	}
	return key
}

// Cache TTL constants
const (
	CandidateListTTL  = 10 * time.Minute
	CandidateItemTTL  = 15 * time.Minute
	ProjectListTTL    = 10 * time.Minute
	ProjectItemTTL    = 15 * time.Minute
	NewsListTTL       = 10 * time.Minute
	NewsItemTTL       = 15 * time.Minute
	VoteCountTTL      = 5 * time.Minute
	UserProfileTTL    = 30 * time.Minute
)
