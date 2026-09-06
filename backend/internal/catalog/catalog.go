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
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Summary   string `json:"summary"`
	Glyph     string `json:"glyph"`
	Cover     int    `json:"cover"`
	CoverURL  string `json:"coverUrl"`
	Badge     string `json:"badge"`
	Author    string `json:"author"`
	Engine    string `json:"engine"`
	Size      string `json:"size"`
	Plays     string `json:"plays"`
	Likes     int64  `json:"likes"`
	Liked     bool   `json:"liked"`
	HasSource bool   `json:"hasSource"`
	Genre     string `json:"genre"`
	License   string `json:"license"`
	Status    string `json:"status"`
	PlayURL   string `json:"playUrl"`
	SourceURL string `json:"sourceUrl"`
	CreatedAt string `json:"createdAt"`
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
	ID         int64  `json:"id"`
	PostID     int64  `json:"postId"`
	Author     string `json:"author"`
	AvatarText string `json:"avatarText"`
	Floor      int    `json:"floor"`
	Content    string `json:"content"`
	CreatedAt  string `json:"createdAt"`
	Likes      int64  `json:"likes"`
	Liked      bool   `json:"liked"`
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
		repos:     []Repo{{ID: 1, Icon: "◆", Name: "void-corridor", Desc: "Roguelike 动作 RPG 完整源码", Lang: "GDScript", Dots: "#478CBF", License: "MIT", Stars: "8.2k", Forks: "1.1k", Downloads: "24.3k", Size: "24 MB", Badge: "完整模板"}},
		replies:   make(map[int64][]Reply),
		gameLikes: make(map[int64]map[int64]bool),
		postLikes: make(map[int64]map[int64]bool),
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

func (s *Store) AddPost(author, title, cat, content string, tags []string, barID int64) (ForumPost, error) {
	title, content = strings.TrimSpace(title), strings.TrimSpace(content)
	if author == "" || title == "" || content == "" || barID <= 0 {
		return ForumPost{}, errors.New("author, title, content and barId are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	item := ForumPost{ID: s.nextPostID, Ava: strings.ToUpper(string([]rune(author)[0])), BG: "linear-gradient(135deg,#06B6D4,#3B82F6)", Name: author, Level: "lv1", Time: "刚刚", Cat: strings.TrimSpace(cat), Tags: append([]string(nil), tags...), Title: title, Excerpt: content, Replies: "0", Views: "0", BarID: barID}
	s.nextPostID++
	s.posts = append(s.posts, item)
	return item, nil
}

func (s *Store) ReviewGame(id int64, status string) (Game, error) {
	status = normalizeGameReviewStatus(status)
	if status == "" {
		return Game{}, errors.New("status must be approved, rejected, published or offline")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for index := range s.games {
		if s.games[index].ID == id {
			if status == "approved" {
				status = "published"
			}
			s.games[index].Status = status
			return s.games[index], nil
		}
	}
	return Game{}, errors.New("game not found")
}

func (s *Store) ReviewPost(id int64, status string) (ForumPost, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	if status != "approved" && status != "hidden" && status != "rejected" {
		return ForumPost{}, errors.New("status must be approved, hidden or rejected")
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

func (s *Store) AddReply(author, content string, postID int64) (Reply, error) {
	content = strings.TrimSpace(content)
	if author == "" || content == "" || postID <= 0 {
		return Reply{}, errors.New("author, content and postId are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for index := range s.posts {
		if s.posts[index].ID != postID {
			continue
		}
		reply := Reply{ID: s.nextReplyID, PostID: postID, Author: author, AvatarText: strings.ToUpper(string([]rune(author)[0])), Floor: len(s.replies[postID]) + 2, Content: content, CreatedAt: "刚刚"}
		s.nextReplyID++
		s.replies[postID] = append(s.replies[postID], reply)
		s.posts[index].Replies = strconv.Itoa(len(s.replies[postID]))
		return reply, nil
	}
	return Reply{}, errors.New("post not found")
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
	return append([]Reply(nil), s.replies[postID]...), true
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
