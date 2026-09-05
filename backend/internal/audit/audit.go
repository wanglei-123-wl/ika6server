package audit

import "github.com/wanglei-123-wl/ika6server/backend/internal/posts"

type Service struct {
	posts *posts.Store
}

func NewService(store *posts.Store) *Service {
	return &Service{posts: store}
}

func (s *Service) Approve(postID int64) (posts.Post, error) {
	return s.posts.SetStatus(postID, posts.StatusApproved)
}

func (s *Service) Reject(postID int64) (posts.Post, error) {
	return s.posts.SetStatus(postID, posts.StatusRejected)
}
