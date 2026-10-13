package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPostsHandler_Routes(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/posts", nil)
	w := httptest.NewRecorder()

	GetPosts(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestFeedHandler_Routes(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/feed", nil)
	w := httptest.NewRecorder()

	GetFeed(w, r)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("expected status 501, got %d", w.Code)
	}
}

func TestPostStruct(t *testing.T) {
	post := Post{
		ID:        "test-id",
		UserID:    "user-123",
		Content:   "test content",
		CreatedAt: "2024-01-01T00:00:00Z",
	}

	if post.ID != "test-id" {
		t.Errorf("expected ID 'test-id', got '%s'", post.ID)
	}
	if post.UserID != "user-123" {
		t.Errorf("expected UserID 'user-123', got '%s'", post.UserID)
	}
}
