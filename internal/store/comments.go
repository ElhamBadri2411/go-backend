package store

import (
	"context"
	"database/sql"
)

type Comment struct {
	ID        int64  `json:"id"`
	Content   string `json:"content"`
	UserId    int64  `json:"user_id"`
	PostId    int64  `json:"post_id"`
	CreatedAt string `json:"created_at"`
	User      User   `json:"user"`
}

type CommentRepositoryPostgres struct {
	db *sql.DB
}

func (s *CommentRepositoryPostgres) GetByPostId(ctx context.Context, postId int64) ([]Comment, error) {
	query := `SELECT c.id, c.post_id, c.user_id, c.content, c.created_at, u.username, u.id
	FROM comments c JOIN users u on u.id = c.user_id 
	WHERE c.post_id = $1 
	ORDER BY c.created_at DESC`

	ctx, cancel := context.WithTimeout(ctx, QueryContextTimeoutDuration)
	defer cancel()
	rows, err := s.db.QueryContext(ctx, query, postId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []Comment

	for rows.Next() {
		var comment Comment
		comment.User = User{}

		err := rows.Scan(&comment.ID, &comment.PostId, &comment.UserId, &comment.Content, &comment.CreatedAt, &comment.User.Username, &comment.User.ID)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, nil
}

// func (s *CommentRepositoryPostgres) Create(ctx context.Context, comment *Comment) error {
// 	query := ` INSERT INTO comments ()`
// 	return nil
// }
