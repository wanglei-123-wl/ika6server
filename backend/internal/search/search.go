package search

import (
	"strings"

	"github.com/wanglei-123-wl/ika6server/backend/internal/posts"
)

func Posts(items []posts.Post, keyword string) []posts.Post {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" {
		return items
	}

	result := make([]posts.Post, 0)
	for _, post := range items {
		text := strings.ToLower(post.Title + " " + post.Description + " " + post.Category)
		if strings.Contains(text, keyword) {
			result = append(result, post)
		}
	}

	return result
}
