package catalog

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

type SQLRepository struct {
	db *sql.DB
}

type GameFile struct {
	ID           int64  `json:"id"`
	GameID       int64  `json:"gameId"`
	Kind         string `json:"kind"`
	OriginalName string `json:"originalName"`
	StoredName   string `json:"storedName"`
	Size         int64  `json:"size"`
	SHA256       string `json:"sha256"`
	DownloadURL  string `json:"downloadUrl"`
	Status       string `json:"status"`
}

func NewSQLRepository(db *sql.DB) (*SQLRepository, error) {
	if db == nil {
		return nil, errors.New("sql database is required")
	}
	return &SQLRepository{db: db}, nil
}

func (r *SQLRepository) Games(ctx context.Context) ([]Game, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, summary, engine, genre, license, status,
		       cover_url, play_url, source_url, plays, likes, created_at
		FROM games
		WHERE status = 'published'
		ORDER BY created_at DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Game, 0)
	for rows.Next() {
		item, err := scanGame(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) Game(ctx context.Context, id int64) (Game, error) {
	return scanGame(r.db.QueryRowContext(ctx, `
		SELECT id, title, summary, engine, genre, license, status,
		       cover_url, play_url, source_url, plays, likes, created_at
		FROM games WHERE id = $1`, id))
}

func (r *SQLRepository) CreateGame(ctx context.Context, ownerID int64, title, summary, description, engine, genre, license string) (Game, error) {
	if ownerID <= 0 || strings.TrimSpace(title) == "" {
		return Game{}, errors.New("owner and title are required")
	}
	return scanGame(r.db.QueryRowContext(ctx, `
		INSERT INTO games (owner_id, title, summary, description, engine, genre, license)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, title, summary, engine, genre, license, status,
		          cover_url, play_url, source_url, plays, likes, created_at`,
		ownerID, strings.TrimSpace(title), strings.TrimSpace(summary), strings.TrimSpace(description),
		strings.TrimSpace(engine), strings.TrimSpace(genre), strings.TrimSpace(license)))
}

func (r *SQLRepository) DeleteGame(ctx context.Context, gameID int64) error {
	if gameID <= 0 {
		return errors.New("game id is required")
	}
	_, err := r.db.ExecContext(ctx, `DELETE FROM games WHERE id = $1`, gameID)
	return err
}

func (r *SQLRepository) LikeGame(ctx context.Context, gameID, userID int64) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `
		INSERT INTO game_likes (game_id, user_id) VALUES ($1, $2)
		ON CONFLICT (game_id, user_id) DO NOTHING`, gameID, userID)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	if err != nil || count == 0 {
		return false, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE games SET likes = likes + 1, updated_at = now() WHERE id = $1`, gameID); err != nil {
		return false, err
	}
	return true, tx.Commit()
}

func (r *SQLRepository) PlayGame(ctx context.Context, gameID int64) error {
	result, err := r.db.ExecContext(ctx, `UPDATE games SET plays = plays + 1, updated_at = now() WHERE id = $1`, gameID)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err == nil && count == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (r *SQLRepository) ReviewGame(ctx context.Context, gameID int64, status string) error {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "approved" {
		status = "published"
	}
	if status != "published" && status != "rejected" && status != "offline" {
		return errors.New("invalid game review status")
	}
	result, err := r.db.ExecContext(ctx, `UPDATE games SET status = $1, updated_at = now() WHERE id = $2`, status, gameID)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err == nil && count == 0 {
		return sql.ErrNoRows
	}
	if err == nil {
		_, err = r.db.ExecContext(ctx, `UPDATE game_files SET status = $1 WHERE game_id = $2`, status, gameID)
	}
	return err
}

func (r *SQLRepository) CreateForumPost(ctx context.Context, barID, authorID int64, title, category, content string, tags []string) (int64, error) {
	if barID <= 0 || authorID <= 0 || strings.TrimSpace(title) == "" || strings.TrimSpace(content) == "" {
		return 0, errors.New("bar, author, title and content are required")
	}
	var id int64
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return 0, err
	}
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO forum_posts (bar_id, author_id, title, category, content, tags)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		barID, authorID, strings.TrimSpace(title), strings.TrimSpace(category), strings.TrimSpace(content), tagsJSON).Scan(&id)
	return id, err
}

func (r *SQLRepository) ForumPosts(ctx context.Context) ([]ForumPost, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id, u.username, p.category, p.title, p.content, p.tags,
		       p.likes, p.views, p.bar_id, p.status
		FROM forum_posts p
		JOIN users u ON u.id = p.author_id
		WHERE p.status = 'published'
		ORDER BY p.created_at DESC, p.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanForumPosts(rows)
}

func (r *SQLRepository) Bars(ctx context.Context) ([]Bar, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT b.id, b.name, b.description,
		       COUNT(DISTINCT p.id), COUNT(DISTINCT p.author_id)
		FROM forum_bars b
		LEFT JOIN forum_posts p ON p.bar_id = b.id AND p.status <> 'hidden'
		GROUP BY b.id, b.name, b.description
		ORDER BY b.id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Bar, 0)
	for rows.Next() {
		var item Bar
		var posts, members int64
		if err := rows.Scan(&item.ID, &item.Name, &item.Desc, &posts, &members); err != nil {
			return nil, err
		}
		item.Icon = "◆"
		item.Posts = formatCount(posts)
		item.Members = formatCount(members)
		item.Hot = posts >= 100
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) ForumPost(ctx context.Context, id int64) (ForumPost, error) {
	var item ForumPost
	var author, content string
	var likes, views int64
	var tagsJSON []byte
	err := r.db.QueryRowContext(ctx, `
		SELECT p.id, u.username, p.category, p.title, p.content, p.tags,
		       p.likes, p.views, p.bar_id, p.status
		FROM forum_posts p
		JOIN users u ON u.id = p.author_id
		WHERE p.id = $1 AND p.status <> 'hidden'`, id).
		Scan(&item.ID, &author, &item.Cat, &item.Title, &content, &tagsJSON, &likes, &views, &item.BarID, &item.Status)
	if err != nil {
		return ForumPost{}, err
	}
	_ = json.Unmarshal(tagsJSON, &item.Tags)
	return forumPostFromRow(item, author, content, likes, views), nil
}

func (r *SQLRepository) Replies(ctx context.Context, postID int64) ([]Reply, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.post_id, u.username, c.content, c.created_at, c.likes
		FROM comments c
		JOIN users u ON u.id = c.author_id
		WHERE c.post_id = $1
		ORDER BY c.created_at ASC, c.id ASC`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Reply, 0)
	floor := 2
	for rows.Next() {
		var item Reply
		var author string
		var createdAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.PostID, &author, &item.Content, &createdAt, &item.Likes); err != nil {
			return nil, err
		}
		item.Author = author
		item.AvatarText = strings.ToUpper(string([]rune(author)[0]))
		item.Floor = floor
		item.CreatedAt = "刚刚"
		if createdAt.Valid {
			item.CreatedAt = createdAt.Time.UTC().Format("2006-01-02T15:04:05Z07:00")
		}
		items = append(items, item)
		floor++
	}
	return items, rows.Err()
}

func (r *SQLRepository) LikeForumPost(ctx context.Context, postID, userID int64) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `
		INSERT INTO forum_post_likes (post_id, user_id) VALUES ($1, $2)
		ON CONFLICT (post_id, user_id) DO NOTHING`, postID, userID)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	if err != nil || count == 0 {
		return false, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE forum_posts SET likes = likes + 1, updated_at = now() WHERE id = $1`, postID); err != nil {
		return false, err
	}
	return true, tx.Commit()
}

func (r *SQLRepository) ReviewForumPost(ctx context.Context, postID int64, status string) error {
	status = strings.ToLower(strings.TrimSpace(status))
	if status != "approved" && status != "hidden" && status != "rejected" {
		return errors.New("invalid post review status")
	}
	result, err := r.db.ExecContext(ctx, `UPDATE forum_posts SET status = $1, updated_at = now() WHERE id = $2`, status, postID)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err == nil && count == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (r *SQLRepository) AddReply(ctx context.Context, postID, authorID int64, content string) (int64, error) {
	if postID <= 0 || authorID <= 0 || strings.TrimSpace(content) == "" {
		return 0, errors.New("post, author and content are required")
	}
	var id int64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO comments (post_id, author_id, content)
		VALUES ($1, $2, $3) RETURNING id`,
		postID, authorID, strings.TrimSpace(content)).Scan(&id)
	return id, err
}

func (r *SQLRepository) Repositories(ctx context.Context) ([]Repo, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, description, language, license, download_url, downloads
		FROM repositories ORDER BY downloads DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Repo, 0)
	for rows.Next() {
		item, err := scanRepo(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) Repository(ctx context.Context, id int64) (Repo, error) {
	return scanRepo(r.db.QueryRowContext(ctx, `
		SELECT id, name, description, language, license, download_url, downloads
		FROM repositories WHERE id = $1`, id))
}

func (r *SQLRepository) DownloadRepository(ctx context.Context, id int64) (Repo, error) {
	return scanRepo(r.db.QueryRowContext(ctx, `
		UPDATE repositories SET downloads = downloads + 1
		WHERE id = $1
		RETURNING id, name, description, language, license, download_url, downloads`, id))
}

func (r *SQLRepository) RecordGameFile(ctx context.Context, gameID int64, file GameFile) error {
	if gameID <= 0 || file.Kind == "" || file.OriginalName == "" || file.StoredName == "" {
		return errors.New("game and file metadata are required")
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO game_files
			(game_id, kind, original_name, stored_name, size, sha256, download_url, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (game_id, kind) DO UPDATE SET
			original_name = EXCLUDED.original_name,
			stored_name = EXCLUDED.stored_name,
			size = EXCLUDED.size,
			sha256 = EXCLUDED.sha256,
			download_url = EXCLUDED.download_url,
			status = EXCLUDED.status`,
		gameID, file.Kind, file.OriginalName, file.StoredName, file.Size, file.SHA256, file.DownloadURL, file.Status)
	return err
}

func (r *SQLRepository) GameFile(ctx context.Context, gameID int64, kind string) (GameFile, error) {
	var file GameFile
	err := r.db.QueryRowContext(ctx, `
		SELECT id, game_id, kind, original_name, stored_name, size, sha256, download_url, status
		FROM game_files
		WHERE game_id = $1 AND kind = $2`,
		gameID, strings.ToLower(strings.TrimSpace(kind))).
		Scan(&file.ID, &file.GameID, &file.Kind, &file.OriginalName, &file.StoredName, &file.Size, &file.SHA256, &file.DownloadURL, &file.Status)
	return file, err
}

func (r *SQLRepository) DownloadGameSource(ctx context.Context, gameID int64) (GameFile, error) {
	var file GameFile
	err := r.db.QueryRowContext(ctx, `
		UPDATE game_files
		SET downloads = downloads + 1
		WHERE game_id = $1 AND kind = 'source' AND status = 'published'
		RETURNING id, game_id, kind, original_name, stored_name, size, sha256, download_url, status`, gameID).
		Scan(&file.ID, &file.GameID, &file.Kind, &file.OriginalName, &file.StoredName, &file.Size, &file.SHA256, &file.DownloadURL, &file.Status)
	return file, err
}

type SearchData struct {
	Games []Game      `json:"games"`
	Posts []ForumPost `json:"posts"`
	Repos []Repo      `json:"repos"`
}

type DevDocs struct {
	Title string   `json:"title"`
	Sub   string   `json:"sub"`
	Steps []string `json:"steps"`
	Code  string   `json:"code"`
}

func (r *SQLRepository) DevDocs(ctx context.Context) (DevDocs, error) {
	var docs DevDocs
	var stepsJSON []byte
	err := r.db.QueryRowContext(ctx, `
		SELECT title, subtitle, steps, code
		FROM developer_docs WHERE id = 1`).
		Scan(&docs.Title, &docs.Sub, &stepsJSON, &docs.Code)
	if err != nil {
		return DevDocs{}, err
	}
	if err := json.Unmarshal(stepsJSON, &docs.Steps); err != nil {
		return DevDocs{}, err
	}
	return docs, nil
}

func (r *SQLRepository) UpdateDevDocs(ctx context.Context, docs DevDocs) error {
	if strings.TrimSpace(docs.Title) == "" {
		return errors.New("developer docs title is required")
	}
	stepsJSON, err := json.Marshal(docs.Steps)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO developer_docs (id, title, subtitle, steps, code)
		VALUES (1, $1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			subtitle = EXCLUDED.subtitle,
			steps = EXCLUDED.steps,
			code = EXCLUDED.code,
			updated_at = now()`,
		strings.TrimSpace(docs.Title), strings.TrimSpace(docs.Sub), stepsJSON, docs.Code)
	return err
}

func (r *SQLRepository) HomeStats(ctx context.Context) (map[string]any, error) {
	var projects, plays, contributors int64
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*), COALESCE(SUM(plays), 0), COUNT(DISTINCT owner_id)
		FROM games WHERE status = 'published'`).Scan(&projects, &plays, &contributors)
	if err != nil {
		return nil, err
	}
	return map[string]any{"projects": projects, "plays": plays, "contributors": contributors, "price": 0}, nil
}

func (r *SQLRepository) AdminStats(ctx context.Context) (map[string]any, error) {
	var pendingGames, pendingPosts int64
	err := r.db.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(*) FROM games WHERE status = 'reviewing'),
			(SELECT COUNT(*) FROM forum_posts WHERE status = 'pending')`).
		Scan(&pendingGames, &pendingPosts)
	if err != nil {
		return nil, err
	}
	return map[string]any{"pendingGames": pendingGames, "pendingPosts": pendingPosts}, nil
}

func (r *SQLRepository) Search(ctx context.Context, keyword string) (SearchData, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return SearchData{Games: []Game{}, Posts: []ForumPost{}, Repos: []Repo{}}, nil
	}
	pattern := "%" + strings.ToLower(keyword) + "%"
	games, err := r.searchGames(ctx, pattern)
	if err != nil {
		return SearchData{}, err
	}
	repos, err := r.searchRepos(ctx, pattern)
	if err != nil {
		return SearchData{}, err
	}
	posts, err := r.searchPosts(ctx, pattern)
	if err != nil {
		return SearchData{}, err
	}
	return SearchData{Games: games, Posts: posts, Repos: repos}, nil
}

func (r *SQLRepository) searchGames(ctx context.Context, pattern string) ([]Game, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, summary, engine, genre, license, status,
		       cover_url, play_url, source_url, plays, likes, created_at
		FROM games
		WHERE status = 'published'
		  AND LOWER(title || ' ' || summary || ' ' || engine || ' ' || genre) LIKE $1
		ORDER BY created_at DESC, id DESC`, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Game, 0)
	for rows.Next() {
		item, err := scanGame(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) searchRepos(ctx context.Context, pattern string) ([]Repo, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, description, language, license, download_url, downloads
		FROM repositories
		WHERE LOWER(name || ' ' || description || ' ' || language) LIKE $1
		ORDER BY downloads DESC, id DESC`, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Repo, 0)
	for rows.Next() {
		item, err := scanRepo(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) searchPosts(ctx context.Context, pattern string) ([]ForumPost, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id, u.username, p.category, p.title, p.content, p.tags,
		       p.likes, p.views, p.bar_id, p.status
		FROM forum_posts p
		JOIN users u ON u.id = p.author_id
		WHERE p.status = 'published'
		  AND LOWER(p.title || ' ' || p.content || ' ' || p.category) LIKE $1
		ORDER BY p.created_at DESC, p.id DESC`, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanForumPosts(rows)
}

func scanForumPosts(rows *sql.Rows) ([]ForumPost, error) {
	items := make([]ForumPost, 0)
	for rows.Next() {
		var item ForumPost
		var author, content string
		var likes, views int64
		var tagsJSON []byte
		if err := rows.Scan(&item.ID, &author, &item.Cat, &item.Title, &content, &tagsJSON, &likes, &views, &item.BarID, &item.Status); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(tagsJSON, &item.Tags)
		items = append(items, forumPostFromRow(item, author, content, likes, views))
	}
	return items, rows.Err()
}

func forumPostFromRow(item ForumPost, author, content string, likes, views int64) ForumPost {
	item.Name = author
	item.Ava = strings.ToUpper(string([]rune(author)[0]))
	item.Level = "lv1"
	item.Time = "刚刚"
	item.Excerpt = content
	item.Likes = likes
	item.Views = formatCount(views)
	return item
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanGame(row rowScanner) (Game, error) {
	var item Game
	var plays, likes int64
	var createdAt sql.NullTime
	err := row.Scan(&item.ID, &item.Title, &item.Summary, &item.Engine, &item.Genre, &item.License,
		&item.Status, &item.CoverURL, &item.PlayURL, &item.SourceURL, &plays, &likes, &createdAt)
	if err != nil {
		return Game{}, err
	}
	item.Likes = likes
	item.Plays = formatCount(plays)
	if createdAt.Valid {
		item.CreatedAt = createdAt.Time.UTC().Format("2006-01-02T15:04:05Z07:00")
	}
	item.HasSource = item.SourceURL != ""
	return item, nil
}

func scanRepo(row rowScanner) (Repo, error) {
	var item Repo
	var downloads int64
	err := row.Scan(&item.ID, &item.Name, &item.Desc, &item.Lang, &item.License, &item.DownloadURL, &downloads)
	if err != nil {
		return Repo{}, err
	}
	item.Icon = "◆"
	item.Downloads = formatCount(downloads)
	return item, nil
}

func formatCount(value int64) string {
	if value < 1000 {
		return strconv.FormatInt(value, 10)
	}
	formatted := strconv.FormatFloat(float64(value)/1000, 'f', 1, 64)
	return strings.TrimSuffix(formatted, ".0") + "k"
}
