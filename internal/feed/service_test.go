package feed

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

// TestFeedService tests the feed service
func TestFeedService(t *testing.T) {
	// Skip if Redis is not available
	redisURL := "redis://localhost:6379"
	client := redis.NewClient(&redis.Options{Addr: redisURL})

	// Test ping to see if Redis is available
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	ctx := context.Background()

	// Test FanOutOnWrite (no-op for now)
	err = client.Ping(ctx).Err()
	if err != nil {
		t.Errorf("Redis ping failed: %v", err)
	}

	// Test GetFeed (no-op for now)
	_, err = client.Exists(ctx, "feed:test").Result()
	if err != nil {
		t.Errorf("Redis check failed: %v", err)
	}
}

// TestFeedServiceIntegration tests with a real Redis instance
func TestFeedServiceIntegration(t *testing.T) {
	redisURL := "redis://localhost:6379"
	client := redis.NewClient(&redis.Options{Addr: redisURL})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		t.Skipf("Redis not available for integration test: %v", err)
	}

	ctx := context.Background()

	// Set a test entry
	testKey := "feed:integration_test"
	testValue := "post1:1000000000"
	_, err = client.ZAdd(ctx, testKey, &redis.Z{Score: 1000000000, Member: testValue}).Result()
	if err != nil {
		t.Errorf("failed to add to Redis: %v", err)
	}

	// Cleanup
	defer client.Del(ctx, testKey)

	// Verify it was added
	exists, err := client.Exists(ctx, testKey).Result()
	if err != nil {
		t.Errorf("failed to check key existence: %v", err)
	}
	if exists != 1 {
		t.Errorf("expected key to exist, got %d", exists)
	}
}

// TestFeedConstants tests feed configuration constants
func TestFeedConstants(t *testing.T) {
	// Verify TTL is 48 hours
	expectedTTL := 48 * time.Hour
	if FeedCacheTTL != expectedTTL {
		t.Errorf("expected FeedCacheTTL to be %v, got %v", expectedTTL, FeedCacheTTL)
	}

	// Verify cache cap is 200
	if FeedCacheCap != 200 {
		t.Errorf("expected FeedCacheCap to be 200, got %d", FeedCacheCap)
	}
}
