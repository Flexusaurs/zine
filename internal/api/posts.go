package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// PostsHandler handles /api/posts routes
func PostsHandler(r chi.Router) {
	r.Get("/", GetPosts)
	r.Post("/", CreatePost)
	r.Get("/{id}", GetPost)
}

// FeedHandler handles /api/feed routes
func FeedHandler(r chi.Router) {
	r.Get("/", GetFeed)
}

// Post represents a post in the system
type Post struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// GetPosts returns paginated posts
func GetPosts(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"posts": []Post{},
		"next":  "",
	})
}

// CreatePost creates a new post
func CreatePost(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.WriteHeader(http.StatusNotImplemented)
}

// GetPost returns a single post by ID
func GetPost(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.WriteHeader(http.StatusNotImplemented)
}

// GetFeed returns the user's feed
func GetFeed(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.WriteHeader(http.StatusNotImplemented)
}
