package feed

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// FeedService handles feed operations
type FeedService struct {
	redis *redis.Client
}

// NewFeedService creates a new feed service
func NewFeedService(redis *redis.Client) *FeedService {
	return &FeedService{redis: redis}
}

// FanOutOnWrite fans out a post to all followers
func (s *FeedService) FanOutOnWrite(ctx context.Context, postID, userID string) error {
	return nil
}

// GetFeed retrieves the feed for a user
func (s *FeedService) GetFeed(ctx context.Context, userID string, cursor string, limit int) ([]string, error) {
	return []string{}, nil
}

const (
	FeedCacheTTL = 48 * time.Hour
	FeedCacheCap = 200
)

func (s *FeedService) cacheKey(userID string) string {
	return "feed:" + userID
}
