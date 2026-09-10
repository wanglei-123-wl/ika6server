package catalog

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Game struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	Summary      string `json:"summary"`
	Glyph        string `json:"glyph"`
	Cover        int    `json:"cover"`
	CoverURL     string `json:"coverUrl"`
	Badge        string `json:"badge"`
	Author       string `json:"author"`
	Engine       string `json:"engine"`
	Size         string `json:"size"`
	Plays        string `json:"plays"`
	Likes        int64  `json:"likes"`
	Liked        bool   `json:"liked"`
	HasSource    bool   `json:"hasSource"`
	Genre        string `json:"genre"`
	License      string `json:"license"`
	Status       string `json:"status"`
	PlayURL      string `json:"playUrl"`
	SourceURL    string `json:"sourceUrl"`
	CreatedAt    string `json:"createdAt"`
	RejectReason string `json:"rejectReason,omitempty"`
}

type DeveloperGame struct {
	ID             int64  `json:"id"`
	Title          string `json:"title"`
	Author         string `json:"author,omitempty"`
	Glyph          string `json:"glyph"`
	Cover          int    `json:"cover"`
	CoverURL       string `json:"coverUrl"`
	Status         string `json:"status"`
	Genre          string `json:"genre"`
	Engine         string `json:"engine"`
	Version        string `json:"version"`
	Plays          int64  `json:"plays"`
	Downloads      int64  `json:"downloads"`
	Likes          int64  `json:"likes"`
	ReviewProgress int    `json:"reviewProgress"`
	ReviewMessage  string `json:"reviewMessage"`
	RejectReason   string `json:"rejectReason"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
	PlayURL        string `json:"playUrl"`
	SourceURL      string `json:"sourceUrl"`
}

type GameUpdate struct {
	Title       *string `json:"title"`
	Summary     *string `json:"summary"`
	Description *string `json:"description"`
	Engine      *string `json:"engine"`
	Genre       *string `json:"genre"`
	License     *string `json:"license"`
	CoverURL    *string `json:"coverUrl"`
}

type AdminForumPost struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	Author     string `json:"author"`
	Content    string `json:"content"`
	Status     string `json:"status"`
	CreatedAt  string `json:"createdAt"`
	ReplyCount int64  `json:"replyCount"`
	Views      int64  `json:"views"`
}

type ForumAttachment struct {
	ID           int64  `json:"id"`
	PostID       int64  `json:"postId"`
	UploaderID   int64  `json:"uploaderId,omitempty"`
	OriginalName string `json:"originalName"`
	Size         int64  `json:"size"`
	DownloadURL  string `json:"downloadUrl"`
	CreatedAt    string `json:"createdAt"`
	StoredName   string `json:"-"`
	SHA256       string `json:"-"`
}

type DeveloperStats struct {
	Following          int64  `json:"following"`
	Followers          int64  `json:"followers"`
	TotalLikes         int64  `json:"totalLikes"`
	Works              int64  `json:"works"`
	TotalPlays         int64  `json:"totalPlays"`
	TotalDownloads     int64  `json:"totalDownloads"`
	SponsorIncome      int64  `json:"sponsorIncome"`
	WeeklyNewFollowers int64  `json:"weeklyNewFollowers"`
	PlayTrend          string `json:"playTrend"`
	DownloadTrend      string `json:"downloadTrend"`
	LikeTrend          string `json:"likeTrend"`
	IncomeTrend        string `json:"incomeTrend"`
}

type Bar struct {
	ID      int64  `json:"id"`
	Icon    string `json:"icon"`
	Name    string `json:"name"`
	Desc    string `json:"desc"`
	Posts   string `json:"posts"`
	Members string `json:"members"`
	Hot     bool   `json:"hot"`
}

type ForumPost struct {
	ID      int64    `json:"id"`
	Ava     string   `json:"ava"`
	BG      string   `json:"bg"`
	Name    string   `json:"name"`
	Level   string   `json:"level"`
	Time    string   `json:"time"`
	Cat     string   `json:"cat"`
	Tags    []string `json:"tags"`
	Title   string   `json:"title"`
	Excerpt string   `json:"excerpt"`
	Media   string   `json:"media"`
	Replies string   `json:"replies"`
	Views   string   `json:"views"`
	Likes   int64    `json:"likes"`
	Liked   bool     `json:"liked"`
	BarID   int64    `json:"barId"`
	Status  string   `json:"status"`
}

type ForumPostFilter struct {
	Cat    string
	BarID  int64
	UserID int64
}

type Repo struct {
	ID          int64  `json:"id"`
	Icon        string `json:"icon"`
	Name        string `json:"name"`
	Desc        string `json:"desc"`
	Lang        string `json:"lang"`
	Dots        string `json:"dots"`
	License     string `json:"license"`
	Stars       string `json:"stars"`
	Forks       string `json:"forks"`
	Downloads   string `json:"downloads"`
	Size        string `json:"size"`
	Badge       string `json:"badge"`
	DownloadURL string `json:"downloadUrl"`
}

type Reply struct {
	ID               int64   `json:"id"`
	PostID           int64   `json:"postId"`
	ParentID         *int64  `json:"parentId"`
	ReplyToCommentID *int64  `json:"replyToCommentId"`
	ReplyToAuthor    string  `json:"replyToAuthor"`
	Author           string  `json:"author"`
	AvatarText       string  `json:"avatarText"`
	Floor            int     `json:"floor"`
	Content          string  `json:"content"`
	CreatedAt        string  `json:"createdAt"`
	Likes            int64   `json:"likes"`
	Liked            bool    `json:"liked"`
	ReplyCount       int64   `json:"replyCount"`
	Replies          []Reply `json:"replies"`
}

type Store struct {
	mu          sync.RWMutex
	nextGameID  int64
	nextPostID  int64
	nextReplyID int64
	games       []Game
	bars        []Bar
	posts       []ForumPost
	repos       []Repo
	replies     map[int64][]Reply
	gameLikes   map[int64]map[int64]bool
	postLikes   map[int64]map[int64]bool
	replyLikes  map[int64]map[int64]bool
	forumFiles  map[int64][]ForumAttachment
	nextFileID  int64
}

func NewStore() *Store {
	now := time.Now().UTC().Format(time.RFC3339)
	return &Store{
		nextGameID: 3, nextPostID: 2, nextReplyID: 1,
		games: []Game{
			{ID: 1, Title: "虚空回廊 · Void Corridor", Summary: "Roguelike 动作 RPG 开源模板", Glyph: "◆", Cover: 1, Badge: "new", Author: "@quietforge", Engine: "Godot 4", Size: "24 MB", Plays: "18.4k", Likes: 2100, HasSource: true, Genre: "Roguelike", License: "MIT", Status: "published", PlayURL: "/games/1/play", SourceURL: "/api/games/1/download-source", CreatedAt: now},
			{ID: 2, Title: "Neon Drift", Summary: "极简霓虹竞速游戏", Glyph: "◇", Cover: 2, Author: "@pixelwave", Engine: "Phaser 3", Size: "8 MB", Plays: "6.7k", Likes: 864, HasSource: true, Genre: "Racing", License: "Apache-2.0", Status: "published", PlayURL: "/games/2/play", SourceURL: "/api/games/2/download-source", CreatedAt: now},
		},
		bars: []Bar{{ID: 1, Icon: "◆", Name: "独立游戏吧", Desc: "独立游戏开发交流 · 开源互助 · 作品发布", Posts: "247,503", Members: "38,412", Hot: true}},
		posts: []ForumPost{
			{ID: 1, Ava: "Q", BG: "linear-gradient(135deg,#06B6D4,#3B82F6)", Name: "quietforge", Level: "lv7", Time: "刚刚", Cat: "作品发布", Tags: []string{"开发日志"}, Title: "【开发日志 #14】虚空回廊终于做完 BOSS 战", Excerpt: "肝了整整 18 天，终于把核心战斗循环打磨完成。", Replies: "142", Views: "3.2k", Likes: 328, BarID: 1, Status: "published"},
		},
		repos:      []Repo{{ID: 1, Icon: "◆", Name: "void-corridor", Desc: "Roguelike 动作 RPG 完整源码", Lang: "GDScript", Dots: "#478CBF", License: "MIT", Stars: "8.2k", Forks: "1.1k", Downloads: "24.3k", Size: "24 MB", Badge: "完整模板"}},
		replies:    make(map[int64][]Reply),
		gameLikes:  make(map[int64]map[int64]bool),
		postLikes:  make(map[int64]map[int64]bool),
		replyLikes: make(map[int64]map[int64]bool),
		forumFiles: make(map[int64][]ForumAttachment),
		nextFileID: 1,
	}
}

func (s *Store) Games() []Game {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Game(nil), s.games...)
}
func (s *Store) Bars() []Bar {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Bar(nil), s.bars...)
}
func (s *Store) Posts() []ForumPost {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]ForumPost(nil), s.posts...)
}

func (s *Store) ForumPosts(filter ForumPostFilter) []ForumPost {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cat := strings.TrimSpace(filter.Cat)
	result := make([]ForumPost, 0, len(s.posts))
	for _, item := range s.posts {
		if item.Status != "published" {
			continue
		}
		if cat != "" && item.Cat != cat {
			continue
		}
		if filter.BarID > 0 && item.BarID != filter.BarID {
			continue
		}
		if filter.UserID > 0 {
			item.Liked = s.postLikes[item.ID][filter.UserID]
		} else {
			item.Liked = false
		}
		result = append(result, item)
	}
	return result
}

func (s *Store) HotPosts() []ForumPost {
	items := s.Posts()
	result := make([]ForumPost, 0, len(items))
	for _, item := range items {
		if item.Status == "published" {
			result = append(result, item)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		return postHeat(result[i]) > postHeat(result[j])
	})
	return result
}

func (s *Store) HomeStats() map[string]any {
	games := s.Games()
	projects, plays := 0, int64(0)
	contributors := make(map[string]struct{})
	for _, item := range games {
		if item.Status != "published" {
			continue
		}
		projects++
		plays += parseCount(item.Plays)
		if item.Author != "" {
			contributors[item.Author] = struct{}{}
		}
	}
	return map[string]any{"projects": projects, "plays": plays, "contributors": len(contributors), "price": 0}
}

func (s *Store) Repos() []Repo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Repo(nil), s.repos...)
}

func (s *Store) Game(id int64) (Game, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, item := range s.games {
		if item.ID == id {
			return item, true
		}
	}
	return Game{}, false
}

func (s *Store) RemoveGame(id int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for index, item := range s.games {
		if item.ID == id {
			s.games = append(s.games[:index], s.games[index+1:]...)
			delete(s.gameLikes, id)
			return
		}
	}
}

func (s *Store) DeveloperGames(username, status string) []DeveloperGame {
	s.mu.RLock()
	defer s.mu.RUnlock()
	author := "@" + strings.TrimSpace(username)
	status = strings.ToLower(strings.TrimSpace(status))
	items := make([]DeveloperGame, 0)
	for _, item := range s.games {
		if item.Author != author {
			continue
		}
		if status != "" && status != "all" && item.Status != status {
			continue
		}
		items = append(items, developerGameFromGame(item, 0))
	}
	return items
}

func (s *Store) DeveloperStats(username string) DeveloperStats {
	games := s.DeveloperGames(username, "all")
	var stats DeveloperStats
	for _, item := range games {
		stats.Works++
		stats.TotalLikes += item.Likes
		stats.TotalPlays += item.Plays
		stats.TotalDownloads += item.Downloads
	}
	return stats
}

func (s *Store) UpdateDeveloperGame(username string, id int64, update GameUpdate) (DeveloperGame, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	author := "@" + strings.TrimSpace(username)
	for index := range s.games {
		item := &s.games[index]
		if item.ID != id {
			continue
		}
		if item.Author != author {
			return DeveloperGame{}, errNotGameOwner
		}
		if item.Status != "draft" && item.Status != "rejected" && item.Status != "offline" {
			return DeveloperGame{}, errInvalidGameState
		}
		if update.Title != nil {
			item.Title = strings.TrimSpace(*update.Title)
			if item.Title == "" {
				return DeveloperGame{}, errors.New("title is required")
			}
		}
		if update.Summary != nil {
			item.Summary = strings.TrimSpace(*update.Summary)
		}
		if update.Engine != nil {
			item.Engine = strings.TrimSpace(*update.Engine)
		}
		if update.Genre != nil {
			item.Genre = strings.TrimSpace(*update.Genre)
		}
		if update.License != nil {
			item.License = strings.TrimSpace(*update.License)
		}
		if update.CoverURL != nil {
			item.CoverURL = strings.TrimSpace(*update.CoverURL)
		}
		return developerGameFromGame(*item, 0), nil
	}
	return DeveloperGame{}, errors.New("game not found")
}

func (s *Store) AdminForumPosts(status string) ([]AdminForumPost, int, error) {
	status = normalizeForumPostStatus(status)
	if status == "" {
		return nil, 0, errors.New("invalid post status")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]AdminForumPost, 0)
	for _, item := range s.posts {
		if status != "all" && item.Status != status {
			continue
		}
		items = append(items, AdminForumPost{
			ID: item.ID, Title: item.Title, Author: item.Name, Content: item.Excerpt,
			Status: item.Status, CreatedAt: item.Time, ReplyCount: parseCount(item.Replies), Views: parseCount(item.Views),
		})
	}
	return items, len(items), nil
}

func (s *Store) AddForumAttachment(postID, uploaderID int64, originalName, storedName string, size int64, sha256 string) ForumAttachment {
	s.mu.Lock()
	defer s.mu.Unlock()
	item := ForumAttachment{
		ID: s.nextFileID, PostID: postID, UploaderID: uploaderID, OriginalName: originalName,
		Size: size, StoredName: storedName, SHA256: sha256,
		DownloadURL: "/api/forum/posts/" + IDString(postID) + "/files/" + IDString(s.nextFileID),
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}
	s.nextFileID++
	s.forumFiles[postID] = append(s.forumFiles[postID], item)
	return item
}

func (s *Store) ForumAttachments(postID int64) []ForumAttachment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]ForumAttachment(nil), s.forumFiles[postID]...)
}

func (s *Store) ForumAttachment(postID, fileID int64) (ForumAttachment, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, item := range s.forumFiles[postID] {
		if item.ID == fileID {
			return item, true
		}
	}
	return ForumAttachment{}, false
}

func (s *Store) ResubmitDeveloperGame(username string, id int64) (DeveloperGame, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	author := "@" + strings.TrimSpace(username)
	for index := range s.games {
		if s.games[index].ID != id {
			continue
		}
		if s.games[index].Author != author {
			return DeveloperGame{}, errNotGameOwner
		}
		if s.games[index].Status != "rejected" {
			return DeveloperGame{}, errInvalidGameState
		}
		s.games[index].Status = "reviewing"
		return developerGameFromGame(s.games[index], 0), nil
	}
	return DeveloperGame{}, errors.New("game not found")
}

func (s *Store) DeleteDeveloperGame(username string, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	author := "@" + strings.TrimSpace(username)
	for index, item := range s.games {
		if item.ID != id {
			continue
		}
		if item.Author != author {
			return errNotGameOwner
		}
		if item.Status != "draft" && item.Status != "rejected" && item.Status != "offline" {
			return errInvalidGameState
		}
		s.games = append(s.games[:index], s.games[index+1:]...)
		delete(s.gameLikes, id)
		return nil
	}
	return errors.New("game not found")
}

func (s *Store) DownloadRepo(id int64) (Repo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for index := range s.repos {
		if s.repos[index].ID == id {
			s.repos[index].Downloads = incrementCount(s.repos[index].Downloads)
			return s.repos[index], nil
		}
	}
	return Repo{}, errors.New("repository not found")
}

func (s *Store) Repo(id int64) (Repo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, item := range s.repos {
		if item.ID == id {
			return item, true
		}
	}
	return Repo{}, false
}

func (s *Store) Post(id int64) (ForumPost, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, item := range s.posts {
		if item.ID == id {
			return item, true
		}
	}
	return ForumPost{}, false
}

func (s *Store) ViewPost(id int64) (ForumPost, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for index := range s.posts {
		if s.posts[index].ID == id {
			s.posts[index].Views = incrementCount(s.posts[index].Views)
			return s.posts[index], true
		}
	}
	return ForumPost{}, false
}

func (s *Store) AddGame(author string, title, summary, engine, genre, license string) (Game, error) {
	title = strings.TrimSpace(title)
	if strings.TrimSpace(author) == "" || title == "" {
		return Game{}, errors.New("author and title are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	item := Game{ID: s.nextGameID, Title: title, Summary: strings.TrimSpace(summary), Glyph: "◆", Author: author, Engine: strings.TrimSpace(engine), Genre: strings.TrimSpace(genre), License: strings.TrimSpace(license), Status: "reviewing", CreatedAt: time.Now().UTC().Format(time.RFC3339)}
	s.nextGameID++
	s.games = append(s.games, item)
	return item, nil
}

func (s *Store) LikeGame(id, userID int64) (Game, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for index := range s.games {
		if s.games[index].ID != id {
			continue
		}
		if s.gameLikes[id] == nil {
			s.gameLikes[id] = make(map[int64]bool)
		}
		if s.gameLikes[id][userID] {
			return s.games[index], false, nil
		}
		s.gameLikes[id][userID] = true
		s.games[index].Likes++
		s.games[index].Liked = true
		return s.games[index], true, nil
	}
	return Game{}, false, errors.New("game not found")
}

func (s *Store) PlayGame(id int64) (Game, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for index := range s.games {
		if s.games[index].ID == id {
			s.games[index].Plays = incrementCount(s.games[index].Plays)
			return s.games[index], nil
		}
	}
	return Game{}, errors.New("game not found")
}

func (s *Store) SetGamePlayURL(id int64, playURL string) error {
	return s.SetGameFileURL(id, "build", playURL)
}

func (s *Store) SetGameFileURL(id int64, kind, fileURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for index := range s.games {
		if s.games[index].ID == id {
			switch strings.ToLower(strings.TrimSpace(kind)) {
			case "cover":
				s.games[index].CoverURL = strings.TrimSpace(fileURL)
			case "build":
				s.games[index].PlayURL = strings.TrimSpace(fileURL)
			case "source":
				s.games[index].SourceURL = strings.TrimSpace(fileURL)
				s.games[index].HasSource = strings.TrimSpace(fileURL) != ""
			}
			return nil
		}
	}
	return errors.New("game not found")
}

func (s *Store) AddPost(author, title, cat, content string, tags []string, barID int64) (ForumPost, error) {
	title, content = strings.TrimSpace(title), strings.TrimSpace(content)
	if author == "" || title == "" || content == "" || barID <= 0 {
		return ForumPost{}, errors.New("author, title, content and barId are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	item := ForumPost{ID: s.nextPostID, Ava: strings.ToUpper(string([]rune(author)[0])), BG: "linear-gradient(135deg,#06B6D4,#3B82F6)", Name: author, Level: "lv1", Time: "刚刚", Cat: strings.TrimSpace(cat), Tags: append([]string(nil), tags...), Title: title, Excerpt: content, Replies: "0", Views: "0", BarID: barID, Status: "published"}
	s.nextPostID++
	s.posts = append(s.posts, item)
	return item, nil
}

func (s *Store) ReviewGame(id int64, status string, reasons ...string) (Game, error) {
	status = normalizeGameReviewStatus(status)
	if status == "" {
		return Game{}, errors.New("status must be approved, rejected, published or offline")
	}
	reason := ""
	if len(reasons) > 0 {
		reason = strings.TrimSpace(reasons[0])
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for index := range s.games {
		if s.games[index].ID == id {
			if status == "approved" {
				status = "published"
			}
			s.games[index].Status = status
			if status == "rejected" {
				s.games[index].RejectReason = reason
			} else {
				s.games[index].RejectReason = ""
			}
			return s.games[index], nil
		}
	}
	return Game{}, errors.New("game not found")
}

func (s *Store) ReviewPost(id int64, status string) (ForumPost, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "approved" {
		status = "published"
	}
	if status != "published" && status != "hidden" && status != "rejected" {
		return ForumPost{}, errors.New("status must be approved, published, hidden or rejected")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for index := range s.posts {
		if s.posts[index].ID == id {
			s.posts[index].Status = status
			return s.posts[index], nil
		}
	}
	return ForumPost{}, errors.New("post not found")
}

func normalizeGameReviewStatus(status string) string {
	status = strings.ToLower(strings.TrimSpace(status))
	switch status {
	case "approved", "published", "rejected", "offline":
		return status
	default:
		return ""
	}
}

func normalizeForumPostStatus(status string) string {
	status = strings.ToLower(strings.TrimSpace(status))
	switch status {
	case "", "all":
		return "all"
	case "approved":
		return "published"
	case "pending", "published", "rejected", "hidden":
		return status
	default:
		return ""
	}
}

func (s *Store) LikePost(id, userID int64) (ForumPost, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for index := range s.posts {
		if s.posts[index].ID != id {
			continue
		}
		if s.postLikes[id] == nil {
			s.postLikes[id] = make(map[int64]bool)
		}
		if s.postLikes[id][userID] {
			return s.posts[index], false, nil
		}
		s.postLikes[id][userID] = true
		s.posts[index].Likes++
		s.posts[index].Liked = true
		return s.posts[index], true, nil
	}
	return ForumPost{}, false, errors.New("post not found")
}

func (s *Store) LikeReply(replyID, userID int64) (Reply, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for postID, replies := range s.replies {
		for index := range replies {
			if replies[index].ID != replyID {
				continue
			}
			if s.replyLikes[replyID] == nil {
				s.replyLikes[replyID] = make(map[int64]bool)
			}
			if s.replyLikes[replyID][userID] {
				return replies[index], false, nil
			}
			s.replyLikes[replyID][userID] = true
			s.replies[postID][index].Likes++
			s.replies[postID][index].Liked = true
			return s.replies[postID][index], true, nil
		}
	}
	return Reply{}, false, errors.New("reply not found")
}

func (s *Store) AddReply(author, content string, postID int64) (Reply, error) {
	return s.AddComment(author, content, postID, nil, nil)
}

func (s *Store) AddComment(author, content string, postID int64, parentID, replyToCommentID *int64) (Reply, error) {
	content = strings.TrimSpace(content)
	if author == "" || content == "" || postID <= 0 {
		return Reply{}, errors.New("author, content and postId are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	postIndex := -1
	for index := range s.posts {
		if s.posts[index].ID == postID {
			postIndex = index
			break
		}
	}
	if postIndex < 0 {
		return Reply{}, errors.New("post not found")
	}
	var replyToAuthor string
	if parentID != nil {
		parent, ok := s.replyByIDLocked(*parentID)
		if !ok || parent.PostID != postID || parent.ParentID != nil {
			return Reply{}, errors.New("parent comment not found")
		}
		targetID := *parentID
		if replyToCommentID != nil {
			targetID = *replyToCommentID
		}
		target, ok := s.replyByIDLocked(targetID)
		if !ok || target.PostID != postID || (target.ID != *parentID && (target.ParentID == nil || *target.ParentID != *parentID)) {
			return Reply{}, errors.New("reply target not found")
		}
		replyToCommentID = &targetID
		replyToAuthor = target.Author
	}
	reply := Reply{
		ID:               s.nextReplyID,
		PostID:           postID,
		ParentID:         cloneInt64(parentID),
		ReplyToCommentID: cloneInt64(replyToCommentID),
		ReplyToAuthor:    replyToAuthor,
		Author:           author,
		AvatarText:       strings.ToUpper(string([]rune(author)[0])),
		Floor:            len(s.replies[postID]) + 2,
		Content:          content,
		CreatedAt:        "刚刚",
		Replies:          []Reply{},
	}
	s.nextReplyID++
	s.replies[postID] = append(s.replies[postID], reply)
	s.posts[postIndex].Replies = strconv.Itoa(len(s.replies[postID]))
	return reply, nil
}

func (s *Store) Replies(postID int64) ([]Reply, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	found := false
	for _, item := range s.posts {
		if item.ID == postID {
			found = true
			break
		}
	}
	if !found {
		return nil, false
	}
	return s.commentsLocked(postID, nil, 0), true
}

func (s *Store) RepliesForUser(postID, userID int64) ([]Reply, bool) {
	items, ok := s.Replies(postID)
	if !ok || userID <= 0 {
		return items, ok
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for index := range items {
		items[index].Liked = s.replyLikes[items[index].ID][userID]
	}
	return items, true
}

func (s *Store) Comments(postID int64) ([]Reply, bool) {
	return s.Replies(postID)
}

func (s *Store) CommentsForUser(postID, userID int64) ([]Reply, bool) {
	return s.RepliesForUser(postID, userID)
}

func (s *Store) CommentReplies(parentID int64) ([]Reply, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	parent, ok := s.replyByIDLocked(parentID)
	if !ok || parent.ParentID != nil {
		return nil, false
	}
	return s.commentsLocked(parent.PostID, &parentID, 0), true
}

func (s *Store) CommentRepliesForUser(parentID, userID int64) ([]Reply, bool) {
	items, ok := s.CommentReplies(parentID)
	if !ok || userID <= 0 {
		return items, ok
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for index := range items {
		items[index].Liked = s.replyLikes[items[index].ID][userID]
	}
	return items, true
}

func (s *Store) replyByIDLocked(id int64) (Reply, bool) {
	for _, replies := range s.replies {
		for _, reply := range replies {
			if reply.ID == id {
				return reply, true
			}
		}
	}
	return Reply{}, false
}

func (s *Store) commentsLocked(postID int64, parentID *int64, userID int64) []Reply {
	items := make([]Reply, 0)
	for _, item := range s.replies[postID] {
		if !sameOptionalInt64(item.ParentID, parentID) {
			continue
		}
		item.ReplyCount = s.replyCountLocked(item.ID)
		item.Replies = []Reply{}
		if userID > 0 {
			item.Liked = s.replyLikes[item.ID][userID]
		}
		items = append(items, item)
	}
	return items
}

func (s *Store) replyCountLocked(parentID int64) int64 {
	var count int64
	for _, replies := range s.replies {
		for _, item := range replies {
			if item.ParentID != nil && *item.ParentID == parentID {
				count++
			}
		}
	}
	return count
}

func sameOptionalInt64(left, right *int64) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func cloneInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func incrementCount(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "1"
	}
	if strings.HasSuffix(strings.ToLower(value), "k") {
		number, err := strconv.ParseFloat(strings.TrimSpace(value[:len(value)-1]), 64)
		if err != nil {
			return value
		}
		return strconv.FormatFloat(number+0.001, 'f', 1, 64) + "k"
	}
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return value
	}
	return strconv.FormatInt(n+1, 10)
}

func postHeat(item ForumPost) int64 {
	return item.Likes*5 + parseCount(item.Views) + parseCount(item.Replies)*3
}

func parseCount(value string) int64 {
	value = strings.ToLower(strings.TrimSpace(strings.ReplaceAll(value, ",", "")))
	if value == "" {
		return 0
	}
	multiplier := float64(1)
	if strings.HasSuffix(value, "k") {
		multiplier = 1000
		value = strings.TrimSpace(strings.TrimSuffix(value, "k"))
	}
	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return int64(number * multiplier)
}

func FilterGames(items []Game, query map[string]string) []Game {
	keyword, genre, engine, license := strings.ToLower(strings.TrimSpace(query["keyword"])), strings.ToLower(strings.TrimSpace(query["genre"])), strings.ToLower(strings.TrimSpace(query["engine"])), strings.ToLower(strings.TrimSpace(query["license"]))
	result := make([]Game, 0, len(items))
	for _, item := range items {
		if item.Status != "published" {
			continue
		}
		text := strings.ToLower(item.Title + " " + item.Summary + " " + item.Author)
		if keyword != "" && !strings.Contains(text, keyword) {
			continue
		}
		if genre != "" && strings.ToLower(item.Genre) != genre {
			continue
		}
		if engine != "" && strings.ToLower(item.Engine) != engine {
			continue
		}
		if license != "" && strings.ToLower(item.License) != license {
			continue
		}
		result = append(result, item)
	}
	if query["sort"] == "likes" {
		sort.Slice(result, func(i, j int) bool { return result[i].Likes > result[j].Likes })
	}
	return result
}

func Search(items []Game, posts []ForumPost, repos []Repo, keyword string) (map[string]any, bool) {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" {
		return map[string]any{"games": []Game{}, "posts": []ForumPost{}, "repos": []Repo{}}, false
	}
	games := make([]Game, 0)
	matchedPosts := make([]ForumPost, 0)
	matchedRepos := make([]Repo, 0)
	for _, item := range items {
		if item.Status != "published" {
			continue
		}
		if strings.Contains(strings.ToLower(item.Title+" "+item.Summary+" "+item.Author), keyword) {
			games = append(games, item)
		}
	}
	for _, item := range posts {
		if item.Status != "published" {
			continue
		}
		if strings.Contains(strings.ToLower(item.Title+" "+item.Excerpt+" "+strings.Join(item.Tags, " ")), keyword) {
			matchedPosts = append(matchedPosts, item)
		}
	}
	for _, item := range repos {
		if strings.Contains(strings.ToLower(item.Name+" "+item.Desc+" "+item.Lang), keyword) {
			matchedRepos = append(matchedRepos, item)
		}
	}
	return map[string]any{"games": games, "posts": matchedPosts, "repos": matchedRepos}, true
}

func IDString(id int64) string { return strconv.FormatInt(id, 10) }

var (
	errNotGameOwner     = errors.New("not game owner")
	errInvalidGameState = errors.New("game status does not allow this operation")
)

func IsNotGameOwner(err error) bool {
	return errors.Is(err, errNotGameOwner)
}

func IsInvalidGameState(err error) bool {
	return errors.Is(err, errInvalidGameState)
}

func ErrNotGameOwner() error {
	return errNotGameOwner
}

func ErrInvalidGameState() error {
	return errInvalidGameState
}

func developerGameFromGame(item Game, downloads int64) DeveloperGame {
	result := DeveloperGame{
		ID:             item.ID,
		Title:          item.Title,
		Glyph:          defaultString(item.Glyph, "◆"),
		Cover:          item.Cover,
		CoverURL:       item.CoverURL,
		Status:         item.Status,
		Genre:          item.Genre,
		Engine:         item.Engine,
		Version:        "v1.0.0",
		Plays:          parseCount(item.Plays),
		Downloads:      downloads,
		Likes:          item.Likes,
		ReviewProgress: reviewProgress(item.Status),
		ReviewMessage:  reviewMessage(item.Status),
		RejectReason:   item.RejectReason,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.CreatedAt,
		PlayURL:        item.PlayURL,
		SourceURL:      item.SourceURL,
	}
	if item.Status != "published" {
		result.PlayURL = ""
		result.SourceURL = ""
	}
	return result
}

func reviewProgress(status string) int {
	switch status {
	case "published":
		return 100
	case "reviewing":
		return 30
	case "rejected":
		return 100
	case "offline":
		return 0
	default:
		return 0
	}
}

func reviewMessage(status string) string {
	switch status {
	case "published":
		return "已发布"
	case "reviewing":
		return "等待审核"
	case "rejected":
		return "审核未通过"
	case "offline":
		return "已下架"
	default:
		return ""
	}
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
