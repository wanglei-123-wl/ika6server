package comments

import "time"

type Comment struct {
	ID        int64     `json:"id"`
	PostID    int64     `json:"postId"`
	AuthorID  int64     `json:"authorId"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}
