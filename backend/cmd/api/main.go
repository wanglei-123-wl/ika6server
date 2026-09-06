package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/wanglei-123-wl/ika6server/backend/internal/admin"
	"github.com/wanglei-123-wl/ika6server/backend/internal/audit"
	"github.com/wanglei-123-wl/ika6server/backend/internal/auth"
	"github.com/wanglei-123-wl/ika6server/backend/internal/blocklist"
	"github.com/wanglei-123-wl/ika6server/backend/internal/catalog"
	"github.com/wanglei-123-wl/ika6server/backend/internal/categories"
	"github.com/wanglei-123-wl/ika6server/backend/internal/config"
	"github.com/wanglei-123-wl/ika6server/backend/internal/database"
	"github.com/wanglei-123-wl/ika6server/backend/internal/files"
	"github.com/wanglei-123-wl/ika6server/backend/internal/posts"
	"github.com/wanglei-123-wl/ika6server/backend/internal/reputation"
	"github.com/wanglei-123-wl/ika6server/backend/internal/sandbox"
	"github.com/wanglei-123-wl/ika6server/backend/internal/scanner"
	"github.com/wanglei-123-wl/ika6server/backend/internal/search"
	"github.com/wanglei-123-wl/ika6server/backend/internal/users"
)

type app struct {
	config     config.Config
	database   *database.Database
	users      users.Repository
	auth       *auth.Service
	posts      *posts.Store
	files      *files.Store
	audit      *audit.Service
	blocklist  *blocklist.Store
	reputation *reputation.Store
	catalog    *catalog.Store
	sqlCatalog *catalog.SQLRepository
	httpServer *http.Server
}

type response map[string]any

func main() {
	cfg := config.Load()
	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	userStore := users.NewStoreWithAdmin(cfg.BootstrapAdminAccount)
	var userRepository users.Repository = userStore
	var sqlCatalog *catalog.SQLRepository
	if cfg.DatabaseURL != "" {
		pingContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := db.Ping(pingContext)
		cancel()
		if err != nil {
			log.Fatal(err)
		}
		migrations, err := database.LoadMigrations(cfg.MigrationsDir)
		if err != nil {
			log.Fatal(err)
		}
		if err := database.ApplyMigrations(context.Background(), db.SQL(), migrations); err != nil {
			log.Fatal(err)
		}
		userRepository, err = users.NewSQLRepository(db.SQL(), cfg.BootstrapAdminAccount)
		if err != nil {
			log.Fatal(err)
		}
		sqlCatalog, err = catalog.NewSQLRepository(db.SQL())
		if err != nil {
			log.Fatal(err)
		}
	}
	postStore := posts.NewStore()
	catalogStore := catalog.NewStore()
	blocklistStore := blocklist.NewStore()
	reputationStore := reputation.NewStore()
	application := &app{
		config:   cfg,
		database: db,
		users:    userRepository,
		auth:     auth.NewService(userRepository, cfg.TokenSecret),
		posts:    postStore,
		files: files.NewStore(cfg.UploadDir, cfg.TempDir, scanner.New(scanner.Config{
			ClamScanBin: cfg.ClamScanBin,
			ClamAVDBDir: cfg.ClamAVDBDir,
			SevenZipBin: cfg.SevenZipBin,
			YaraBin:     cfg.YaraBin,
			YaraRules:   cfg.YaraRules,
		}), sandbox.NewAnalyzer(), blocklistStore),
		audit:      audit.NewService(postStore),
		blocklist:  blocklistStore,
		reputation: reputationStore,
		catalog:    catalogStore,
		sqlCatalog: sqlCatalog,
	}

	mux := http.NewServeMux()
	application.routes(mux)

	application.httpServer = &http.Server{
		Addr:    cfg.Addr,
		Handler: withCORS(mux),
	}

	log.Printf("ika6 backend listening on %s", cfg.Addr)
	if err := application.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func (a *app) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/health", a.handleHealth)
	mux.HandleFunc("POST /api/auth/register", a.handleRegister)
	mux.HandleFunc("POST /api/auth/login", a.handleLogin)
	mux.HandleFunc("POST /api/auth/logout", a.handleLogout)
	mux.HandleFunc("GET /api/auth/me", a.handleMe)
	mux.HandleFunc("POST /api/auth/social-login", a.handleSocialLogin)
	mux.HandleFunc("GET /api/users/me", a.handleMe)
	mux.HandleFunc("GET /api/home", a.handleHome)
	mux.HandleFunc("GET /api/games", a.handleGames)
	mux.HandleFunc("POST /api/games", a.handleCreateGame)
	mux.HandleFunc("GET /api/games/{id}", a.handleGame)
	mux.HandleFunc("POST /api/games/{id}/like", a.handleLikeGame)
	mux.HandleFunc("POST /api/games/{id}/play", a.handlePlayGame)
	mux.HandleFunc("GET /api/games/{id}/download-source", a.handleGameSource)
	mux.HandleFunc("GET /api/games/{id}/files/{kind}", a.handleGameFile)
	mux.HandleFunc("GET /api/forum/bars", a.handleForumBars)
	mux.HandleFunc("GET /api/forum/posts", a.handleForumPosts)
	mux.HandleFunc("GET /api/forum/posts/{id}", a.handleForumPost)
	mux.HandleFunc("POST /api/forum/posts", a.handleCreateForumPost)
	mux.HandleFunc("POST /api/forum/posts/{id}/like", a.handleLikeForumPost)
	mux.HandleFunc("GET /api/forum/posts/{id}/replies", a.handleReplies)
	mux.HandleFunc("POST /api/forum/posts/{id}/replies", a.handleCreateReply)
	mux.HandleFunc("GET /api/repos", a.handleRepos)
	mux.HandleFunc("GET /api/repos/{id}", a.handleRepo)
	mux.HandleFunc("GET /api/repos/{id}/download", a.handleRepoDownload)
	mux.HandleFunc("GET /api/dev-docs", a.handleDevDocs)
	mux.HandleFunc("PUT /api/admin/dev-docs", a.handleUpdateDevDocs)
	mux.HandleFunc("GET /api/users/me/reputation", a.handleMyReputation)
	mux.HandleFunc("GET /api/categories", a.handleCategories)
	mux.HandleFunc("GET /api/posts", a.handleListPosts)
	mux.HandleFunc("POST /api/posts", a.handleCreatePost)
	mux.HandleFunc("GET /api/search", a.handleSearch)
	mux.HandleFunc("POST /api/posts/{id}/files", a.handleUploadFile)
	mux.HandleFunc("GET /api/posts/{id}/download", a.handleDownloadFile)
	mux.HandleFunc("POST /api/admin/posts/{id}/approve", a.handleApprovePost)
	mux.HandleFunc("POST /api/admin/posts/{id}/reject", a.handleRejectPost)
	mux.HandleFunc("GET /api/admin/blocklist", a.handleListBlocklist)
	mux.HandleFunc("POST /api/admin/blocklist", a.handleAddBlocklist)
	mux.HandleFunc("GET /api/admin/reputation", a.handleListReputation)
	mux.HandleFunc("GET /api/admin/dashboard", a.handleAdminDashboard)
	mux.HandleFunc("POST /api/admin/games/{id}/review", a.handleReviewGame)
	mux.HandleFunc("POST /api/admin/posts/{id}/review", a.handleReviewForumPost)
	mux.HandleFunc("POST /api/admin/users/{id}/ban", a.handleBanUser)
}

func (a *app) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	if err := a.database.Ping(r.Context()); err != nil {
		status = "degraded"
	}

	writeJSON(w, http.StatusOK, response{
		"ok":      status == "ok",
		"service": "ika6-backend",
		"status":  status,
	})
}

func (a *app) handleRegister(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Name     string `json:"name"`
		Account  string `json:"account"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}

	username := input.Username
	if username == "" {
		username = input.Name
	}
	email := input.Email
	if email == "" {
		email = input.Account
	}
	user, token, err := a.auth.Register(username, email, input.Password)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	a.reputation.Add(user.ID, reputation.EventRegister, 5, "registered account")

	writeAPI(w, http.StatusCreated, map[string]any{"user": publicUser(user, "register"), "token": token})
}

func (a *app) handleMyReputation(w http.ResponseWriter, r *http.Request) {
	user, ok := a.currentUser(w, r)
	if !ok {
		return
	}

	writeJSON(w, http.StatusOK, response{
		"ok":         true,
		"reputation": a.reputation.Get(user.ID),
	})
}

func (a *app) handleLogin(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Account  string `json:"account"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Remember bool   `json:"remember"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}

	email := input.Email
	if email == "" {
		email = input.Account
	}
	user, token, err := a.auth.LoginWithRemember(email, input.Password, input.Remember)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	writeAPI(w, http.StatusOK, map[string]any{"user": publicUser(user, "password"), "token": token})
}

func (a *app) handleMe(w http.ResponseWriter, r *http.Request) {
	user, ok := a.currentUser(w, r)
	if !ok {
		return
	}

	writeAPI(w, http.StatusOK, map[string]any{"user": publicUser(user, "password")})
}

func (a *app) handleLogout(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.currentUser(w, r); !ok {
		return
	}
	writeAPI(w, http.StatusOK, map[string]any{"loggedOut": true})
}

func (a *app) handleSocialLogin(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "social login is not implemented in phase 1")
}

func (a *app) handleHome(w http.ResponseWriter, r *http.Request) {
	if a.sqlCatalog != nil {
		stats, err := a.sqlCatalog.HomeStats(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load home stats")
			return
		}
		games, err := a.sqlCatalog.Games(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load latest games")
			return
		}
		posts, err := a.sqlCatalog.ForumPosts(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load home posts")
			return
		}
		writeAPI(w, http.StatusOK, map[string]any{"stats": stats, "latestGames": games, "hotPosts": posts, "feedPosts": posts})
		return
	}
	writeAPI(w, http.StatusOK, map[string]any{
		"stats":       map[string]any{"projects": len(a.catalog.Games()), "plays": 0, "contributors": 0, "price": 0},
		"latestGames": catalog.FilterGames(a.catalog.Games(), nil),
		"hotPosts":    publishedPosts(a.catalog.Posts()),
		"feedPosts":   publishedPosts(a.catalog.Posts()),
	})
}

func (a *app) handleDevDocs(w http.ResponseWriter, r *http.Request) {
	if a.sqlCatalog != nil {
		docs, err := a.sqlCatalog.DevDocs(r.Context())
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "developer docs not configured")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load developer docs")
			return
		}
		writeAPI(w, http.StatusOK, map[string]any{"quickstart": docs})
		return
	}
	writeAPI(w, http.StatusOK, map[string]any{
		"quickstart": map[string]any{
			"title": "快速开始",
			"sub":   "从 0 到 1，把游戏发布到平台",
			"steps": []string{"注册并登录", "上传游戏文件", "写介绍", "一键发布"},
			"code":  "npm install...",
		},
	})
}

func (a *app) handleUpdateDevDocs(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.currentAdmin(w, r); !ok {
		return
	}
	if a.sqlCatalog == nil {
		writeError(w, http.StatusNotImplemented, "developer docs persistence requires database mode")
		return
	}
	var docs catalog.DevDocs
	if !decodeJSON(w, r, &docs) {
		return
	}
	if err := a.sqlCatalog.UpdateDevDocs(r.Context(), docs); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeAPI(w, http.StatusOK, map[string]any{"updated": true, "quickstart": docs})
}

func (a *app) handleGames(w http.ResponseWriter, r *http.Request) {
	if a.sqlCatalog != nil {
		items, err := a.sqlCatalog.Games(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load games")
			return
		}
		writeAPI(w, http.StatusOK, paginate(catalog.FilterGames(items, map[string]string{
			"keyword": r.URL.Query().Get("keyword"), "genre": r.URL.Query().Get("genre"),
			"engine": r.URL.Query().Get("engine"), "license": r.URL.Query().Get("license"), "sort": r.URL.Query().Get("sort"),
		}), r))
		return
	}
	items := catalog.FilterGames(a.catalog.Games(), map[string]string{
		"keyword": r.URL.Query().Get("keyword"), "genre": r.URL.Query().Get("genre"),
		"engine": r.URL.Query().Get("engine"), "license": r.URL.Query().Get("license"), "sort": r.URL.Query().Get("sort"),
	})
	writeAPI(w, http.StatusOK, paginate(items, r))
}

func (a *app) handleCreateGame(w http.ResponseWriter, r *http.Request) {
	user, ok := a.currentUser(w, r)
	if !ok {
		return
	}
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}
	var (
		item catalog.Game
		err  error
	)
	if a.sqlCatalog != nil {
		item, err = a.sqlCatalog.CreateGame(r.Context(), user.ID, r.FormValue("title"), r.FormValue("summary"), r.FormValue("description"), r.FormValue("engine"), r.FormValue("genre"), r.FormValue("license"))
	} else {
		item, err = a.catalog.AddGame("@"+user.Username, r.FormValue("title"), r.FormValue("summary"), r.FormValue("engine"), r.FormValue("genre"), r.FormValue("license"))
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	uploaded := make(map[string]any)
	for _, field := range []struct {
		formName string
		kind     string
	}{
		{formName: "coverFile", kind: "cover"},
		{formName: "buildFile", kind: "build"},
		{formName: "sourceFile", kind: "source"},
	} {
		file, header, fileErr := r.FormFile(field.formName)
		if fileErr != nil {
			if errors.Is(fileErr, http.ErrMissingFile) {
				continue
			}
			a.removeCreatedGame(item.ID)
			writeError(w, http.StatusBadRequest, "invalid "+field.formName)
			return
		}
		_ = file.Close()
		stored, saveErr := a.files.SaveKind(r.Context(), item.ID, field.kind, header)
		if saveErr != nil {
			a.removeCreatedGame(item.ID)
			writeError(w, http.StatusUnprocessableEntity, field.formName+": "+saveErr.Error())
			return
		}
		if a.sqlCatalog != nil {
			if recordErr := a.sqlCatalog.RecordGameFile(r.Context(), item.ID, catalog.GameFile{
				GameID: item.ID, Kind: field.kind, OriginalName: stored.OriginalName,
				StoredName: stored.StoredName, Size: stored.Size, SHA256: stored.SHA256,
				DownloadURL: "/api/games/" + strconv.FormatInt(item.ID, 10) + "/files/" + field.kind,
				Status:      item.Status,
			}); recordErr != nil {
				a.removeCreatedGame(item.ID)
				writeError(w, http.StatusInternalServerError, "failed to record uploaded file")
				return
			}
		}
		uploaded[field.kind] = stored
	}
	writeAPI(w, http.StatusCreated, map[string]any{"id": item.ID, "title": item.Title, "status": item.Status, "message": "Game submitted for review", "files": uploaded})
}

func (a *app) removeCreatedGame(id int64) {
	if a.files != nil {
		a.files.RemoveOwner(id)
	}
	if a.sqlCatalog != nil {
		_ = a.sqlCatalog.DeleteGame(context.Background(), id)
		return
	}
	if a.catalog != nil {
		a.catalog.RemoveGame(id)
	}
}

func (a *app) handleGame(w http.ResponseWriter, r *http.Request) {
	id, ok := routeID(w, r, "games")
	if !ok {
		return
	}
	if a.sqlCatalog != nil {
		item, err := a.sqlCatalog.Game(r.Context(), id)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "game not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load game")
			return
		}
		writeAPI(w, http.StatusOK, item)
		return
	}
	item, exists := a.catalog.Game(id)
	if !exists {
		writeError(w, http.StatusNotFound, "game not found")
		return
	}
	writeAPI(w, http.StatusOK, item)
}

func (a *app) handleGameSource(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.currentUser(w, r); !ok {
		return
	}
	id, ok := routeID(w, r, "games")
	if !ok {
		return
	}
	if a.sqlCatalog != nil {
		file, err := a.sqlCatalog.DownloadGameSource(r.Context(), id)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "source not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load source")
			return
		}
		writeAPI(w, http.StatusOK, map[string]any{"downloadUrl": file.DownloadURL, "size": file.Size})
		return
	}
	item, exists := a.catalog.Game(id)
	if a.sqlCatalog != nil {
		var err error
		item, err = a.sqlCatalog.Game(r.Context(), id)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "source not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load source")
			return
		}
		exists = true
	}
	if !exists || !item.HasSource {
		writeError(w, http.StatusNotFound, "source not found")
		return
	}
	writeAPI(w, http.StatusOK, map[string]any{"downloadUrl": item.SourceURL, "size": item.Size})
}

func (a *app) handleGameFile(w http.ResponseWriter, r *http.Request) {
	id, ok := routeID(w, r, "games")
	if !ok {
		return
	}
	kind := strings.ToLower(strings.TrimSpace(r.PathValue("kind")))
	if kind != "cover" && kind != "build" && kind != "source" {
		writeError(w, http.StatusBadRequest, "invalid file kind")
		return
	}
	var (
		file   files.File
		path   string
		exists bool
	)
	if a.sqlCatalog != nil {
		record, err := a.sqlCatalog.GameFile(r.Context(), id, kind)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "file not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load file")
			return
		}
		file = files.File{
			ID: record.ID, PostID: record.GameID, OriginalName: record.OriginalName,
			StoredName: record.StoredName, Size: record.Size, SHA256: record.SHA256,
		}
		path = filepath.Join(a.config.UploadDir, record.StoredName)
		_, err = os.Stat(path)
		exists = err == nil
	} else {
		file, path, exists = a.files.FindByKey(id, kind)
	}
	if !exists {
		writeError(w, http.StatusNotFound, "file not found")
		return
	}
	if entry, blocked := a.blocklist.Contains(file.SHA256); blocked {
		writeError(w, http.StatusForbidden, "download blocked: "+entry.Reason)
		return
	}
	w.Header().Set("Content-Disposition", `attachment; filename="`+sanitizeHeader(file.OriginalName)+`"`)
	http.ServeFile(w, r, path)
}

func (a *app) handleLikeGame(w http.ResponseWriter, r *http.Request) {
	user, ok := a.currentUser(w, r)
	if !ok {
		return
	}
	id, ok := routeID(w, r, "games")
	if !ok {
		return
	}
	item, changed, err := a.catalog.LikeGame(id, user.ID)
	if a.sqlCatalog != nil {
		changed, err = a.sqlCatalog.LikeGame(r.Context(), id, user.ID)
		if err == nil {
			item, err = a.sqlCatalog.Game(r.Context(), id)
		}
	}
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeAPI(w, http.StatusOK, map[string]any{"liked": item.Liked, "likes": item.Likes, "changed": changed})
}

func (a *app) handlePlayGame(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.currentUser(w, r); !ok {
		return
	}
	id, ok := routeID(w, r, "games")
	if !ok {
		return
	}
	item, err := a.catalog.PlayGame(id)
	if a.sqlCatalog != nil {
		err = a.sqlCatalog.PlayGame(r.Context(), id)
		if err == nil {
			item, err = a.sqlCatalog.Game(r.Context(), id)
		}
	}
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeAPI(w, http.StatusOK, map[string]any{"id": item.ID, "plays": item.Plays, "playUrl": item.PlayURL})
}

func (a *app) handleForumBars(w http.ResponseWriter, r *http.Request) {
	if a.sqlCatalog != nil {
		items, err := a.sqlCatalog.Bars(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load forum bars")
			return
		}
		writeAPI(w, http.StatusOK, paginate(items, r))
		return
	}
	writeAPI(w, http.StatusOK, paginate(a.catalog.Bars(), r))
}

func (a *app) handleForumPosts(w http.ResponseWriter, r *http.Request) {
	if a.sqlCatalog != nil {
		items, err := a.sqlCatalog.ForumPosts(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load forum posts")
			return
		}
		writeAPI(w, http.StatusOK, paginate(items, r))
		return
	}
	writeAPI(w, http.StatusOK, paginate(publishedPosts(a.catalog.Posts()), r))
}

func (a *app) handleForumPost(w http.ResponseWriter, r *http.Request) {
	id, ok := routeID(w, r, "posts")
	if !ok {
		return
	}
	item, exists := a.catalog.Post(id)
	if a.sqlCatalog != nil {
		item, err := a.sqlCatalog.ForumPost(r.Context(), id)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "post not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load post")
			return
		}
		writeAPI(w, http.StatusOK, item)
		return
	}
	if !exists {
		writeError(w, http.StatusNotFound, "post not found")
		return
	}
	writeAPI(w, http.StatusOK, item)
}

func (a *app) handleCreateForumPost(w http.ResponseWriter, r *http.Request) {
	user, ok := a.currentUser(w, r)
	if !ok {
		return
	}
	var input struct {
		Title   string   `json:"title"`
		Cat     string   `json:"cat"`
		Tags    []string `json:"tags"`
		Content string   `json:"content"`
		BarID   int64    `json:"barId"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	var (
		item catalog.ForumPost
		err  error
	)
	if a.sqlCatalog != nil {
		id, createErr := a.sqlCatalog.CreateForumPost(r.Context(), input.BarID, user.ID, input.Title, input.Cat, input.Content, input.Tags)
		if createErr != nil {
			writeError(w, http.StatusBadRequest, createErr.Error())
			return
		}
		item, err = a.sqlCatalog.ForumPost(r.Context(), id)
	} else {
		item, err = a.catalog.AddPost(user.Username, input.Title, input.Cat, input.Content, input.Tags, input.BarID)
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeAPI(w, http.StatusCreated, item)
}

func (a *app) handleLikeForumPost(w http.ResponseWriter, r *http.Request) {
	user, ok := a.currentUser(w, r)
	if !ok {
		return
	}
	id, ok := routeID(w, r, "posts")
	if !ok {
		return
	}
	item, changed, err := a.catalog.LikePost(id, user.ID)
	if a.sqlCatalog != nil {
		changed, err = a.sqlCatalog.LikeForumPost(r.Context(), id, user.ID)
		if err == nil {
			item, err = a.sqlCatalog.ForumPost(r.Context(), id)
		}
	}
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeAPI(w, http.StatusOK, map[string]any{"liked": item.Liked, "likes": item.Likes, "changed": changed})
}

func (a *app) handleReplies(w http.ResponseWriter, r *http.Request) {
	id, ok := routeID(w, r, "posts")
	if !ok {
		return
	}
	items, exists := a.catalog.Replies(id)
	if a.sqlCatalog != nil {
		sqlItems, err := a.sqlCatalog.Replies(r.Context(), id)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "post not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load replies")
			return
		}
		writeAPI(w, http.StatusOK, paginate(sqlItems, r))
		return
	}
	if !exists {
		writeError(w, http.StatusNotFound, "post not found")
		return
	}
	writeAPI(w, http.StatusOK, paginate(items, r))
}

func (a *app) handleCreateReply(w http.ResponseWriter, r *http.Request) {
	user, ok := a.currentUser(w, r)
	if !ok {
		return
	}
	id, ok := routeID(w, r, "posts")
	if !ok {
		return
	}
	var input struct {
		Content string `json:"content"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	item, err := a.catalog.AddReply(user.Username, input.Content, id)
	if a.sqlCatalog != nil {
		_, err = a.sqlCatalog.AddReply(r.Context(), id, user.ID, input.Content)
		if err == nil {
			items, listErr := a.sqlCatalog.Replies(r.Context(), id)
			if listErr == nil && len(items) > 0 {
				item = items[len(items)-1]
			} else {
				err = listErr
			}
		}
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeAPI(w, http.StatusCreated, item)
}

func (a *app) handleRepos(w http.ResponseWriter, r *http.Request) {
	if a.sqlCatalog != nil {
		items, err := a.sqlCatalog.Repositories(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load repositories")
			return
		}
		writeAPI(w, http.StatusOK, paginate(items, r))
		return
	}
	writeAPI(w, http.StatusOK, paginate(a.catalog.Repos(), r))
}

func (a *app) handleRepo(w http.ResponseWriter, r *http.Request) {
	id, ok := routeID(w, r, "repos")
	if !ok {
		return
	}
	if a.sqlCatalog != nil {
		item, err := a.sqlCatalog.Repository(r.Context(), id)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "repository not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load repository")
			return
		}
		writeAPI(w, http.StatusOK, item)
		return
	}
	item, exists := a.catalog.Repo(id)
	if !exists {
		writeError(w, http.StatusNotFound, "repository not found")
		return
	}
	writeAPI(w, http.StatusOK, item)
}

func (a *app) handleRepoDownload(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.currentUser(w, r); !ok {
		return
	}
	id, ok := routeID(w, r, "repos")
	if !ok {
		return
	}
	if a.sqlCatalog != nil {
		item, err := a.sqlCatalog.DownloadRepository(r.Context(), id)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "repository not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load repository")
			return
		}
		writeAPI(w, http.StatusOK, map[string]any{"downloadUrl": item.DownloadURL, "size": item.Size})
		return
	}
	item, exists := a.catalog.Repo(id)
	if !exists {
		writeError(w, http.StatusNotFound, "repository not found")
		return
	}
	writeAPI(w, http.StatusOK, map[string]any{"downloadUrl": item.DownloadURL, "size": item.Size})
}

func (a *app) handleCategories(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, response{
		"ok":    true,
		"items": categories.List(),
	})
}

func (a *app) handleListPosts(w http.ResponseWriter, r *http.Request) {
	status := posts.Status(r.URL.Query().Get("status"))
	writeJSON(w, http.StatusOK, response{
		"ok":    true,
		"items": a.posts.List(status),
	})
}

func (a *app) handleCreatePost(w http.ResponseWriter, r *http.Request) {
	user, ok := a.currentUser(w, r)
	if !ok {
		return
	}

	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Category    string `json:"category"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}

	post, err := a.posts.Create(user.ID, input.Title, input.Description, input.Category)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, response{
		"ok":   true,
		"item": post,
	})
}

func (a *app) handleSearch(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	if keyword == "" {
		keyword = r.URL.Query().Get("q")
	}
	if a.sqlCatalog != nil {
		result, err := a.sqlCatalog.Search(r.Context(), keyword)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "search failed")
			return
		}
		writeAPI(w, http.StatusOK, result)
		return
	}
	result, _ := catalog.Search(a.catalog.Games(), a.catalog.Posts(), a.catalog.Repos(), keyword)
	result["legacyPosts"] = search.Posts(a.posts.List(posts.StatusApproved), keyword)
	writeAPI(w, http.StatusOK, result)
}

func (a *app) handleUploadFile(w http.ResponseWriter, r *http.Request) {
	user, ok := a.currentUser(w, r)
	if !ok {
		return
	}

	postID, ok := postIDFromPath(w, r)
	if !ok {
		return
	}

	post, exists := a.posts.Find(postID)
	if !exists {
		writeError(w, http.StatusNotFound, "post not found")
		return
	}
	if post.AuthorID != user.ID && !admin.CanManage(user) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	if err := r.ParseMultipartForm(64 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid upload form")
		return
	}

	src, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	_ = src.Close()

	file, err := a.files.Save(r.Context(), postID, header)
	if err != nil {
		a.reputation.Add(user.ID, reputation.EventUploadRejected, -20, err.Error())
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	a.reputation.Add(user.ID, reputation.EventUploadClean, 10, "uploaded file passed scanning")

	writeJSON(w, http.StatusCreated, response{
		"ok":   true,
		"file": file,
	})
}

func (a *app) handleDownloadFile(w http.ResponseWriter, r *http.Request) {
	postID, ok := postIDFromPath(w, r)
	if !ok {
		return
	}

	post, exists := a.posts.Find(postID)
	if !exists || post.Status != posts.StatusApproved {
		writeError(w, http.StatusNotFound, "file not found")
		return
	}

	file, filePath, exists := a.files.FindByPost(postID)
	if !exists {
		writeError(w, http.StatusNotFound, "file not found")
		return
	}
	if entry, blocked := a.blocklist.Contains(file.SHA256); blocked {
		writeError(w, http.StatusForbidden, "download blocked: "+entry.Reason)
		return
	}

	if _, err := a.posts.AddDownload(postID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	a.reputation.Add(post.AuthorID, reputation.EventDownload, 1, "source file downloaded")

	w.Header().Set("Content-Disposition", `attachment; filename="`+sanitizeHeader(file.OriginalName)+`"`)
	http.ServeFile(w, r, filePath)
}

func (a *app) handleApprovePost(w http.ResponseWriter, r *http.Request) {
	a.auditPost(w, r, true)
}

func (a *app) handleRejectPost(w http.ResponseWriter, r *http.Request) {
	a.auditPost(w, r, false)
}

func (a *app) auditPost(w http.ResponseWriter, r *http.Request, approved bool) {
	user, ok := a.currentUser(w, r)
	if !ok {
		return
	}
	if !admin.CanManage(user) {
		writeError(w, http.StatusForbidden, "admin role required")
		return
	}

	postID, ok := postIDFromPath(w, r)
	if !ok {
		return
	}

	var (
		post posts.Post
		err  error
	)
	if approved {
		post, err = a.audit.Approve(postID)
	} else {
		post, err = a.audit.Reject(postID)
	}
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, response{
		"ok":   true,
		"item": post,
	})
	if approved {
		a.reputation.Add(post.AuthorID, reputation.EventPostApproved, 20, "post approved")
	} else {
		a.reputation.Add(post.AuthorID, reputation.EventPostRejected, -30, "post rejected")
	}
}

func (a *app) handleListBlocklist(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.currentAdmin(w, r); !ok {
		return
	}

	writeJSON(w, http.StatusOK, response{
		"ok":    true,
		"items": a.blocklist.List(),
	})
}

func (a *app) handleAddBlocklist(w http.ResponseWriter, r *http.Request) {
	user, ok := a.currentAdmin(w, r)
	if !ok {
		return
	}

	var input struct {
		SHA256 string `json:"sha256"`
		Reason string `json:"reason"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}

	entry, err := a.blocklist.Add(input.SHA256, input.Reason, user.ID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, response{
		"ok":    true,
		"entry": entry,
	})
}

func (a *app) handleListReputation(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.currentAdmin(w, r); !ok {
		return
	}

	writeJSON(w, http.StatusOK, response{
		"ok":    true,
		"items": a.reputation.List(),
	})
}

func (a *app) handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.currentAdmin(w, r); !ok {
		return
	}
	if a.sqlCatalog != nil {
		stats, err := a.sqlCatalog.AdminStats(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load admin stats")
			return
		}
		_, activeUsers := a.users.Counts(time.Now().UTC())
		stats["activeUsers"] = activeUsers
		stats["reports"] = 0
		writeAPI(w, http.StatusOK, stats)
		return
	}
	_, activeUsers := a.users.Counts(time.Now().UTC())
	pendingGames, pendingPosts := 0, 0
	for _, item := range a.catalog.Games() {
		if item.Status == "reviewing" {
			pendingGames++
		}
	}
	for _, item := range a.catalog.Posts() {
		if item.Status == "pending" {
			pendingPosts++
		}
	}
	writeAPI(w, http.StatusOK, map[string]any{
		"pendingGames": pendingGames,
		"pendingPosts": pendingPosts,
		"activeUsers":  activeUsers,
		"reports":      0,
	})
}

func (a *app) handleReviewGame(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.currentAdmin(w, r); !ok {
		return
	}
	id, ok := routeID(w, r, "games")
	if !ok {
		return
	}
	var input struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	var item catalog.Game
	var err error
	if a.sqlCatalog != nil {
		err = a.sqlCatalog.ReviewGame(r.Context(), id, input.Status)
		if err == nil {
			item, err = a.sqlCatalog.Game(r.Context(), id)
		}
	} else {
		item, err = a.catalog.ReviewGame(id, input.Status)
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeAPI(w, http.StatusOK, map[string]any{"item": item, "reason": input.Reason})
}

func (a *app) handleReviewForumPost(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.currentAdmin(w, r); !ok {
		return
	}
	id, ok := routeID(w, r, "posts")
	if !ok {
		return
	}
	var input struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	var item catalog.ForumPost
	var err error
	if a.sqlCatalog != nil {
		err = a.sqlCatalog.ReviewForumPost(r.Context(), id, input.Status)
		if err == nil {
			item, err = a.sqlCatalog.ForumPost(r.Context(), id)
		}
	} else {
		item, err = a.catalog.ReviewPost(id, input.Status)
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeAPI(w, http.StatusOK, map[string]any{"item": item, "reason": input.Reason})
}

func (a *app) handleBanUser(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.currentAdmin(w, r); !ok {
		return
	}
	id, ok := routeID(w, r, "users")
	if !ok {
		return
	}
	var input struct {
		Reason string    `json:"reason"`
		Until  time.Time `json:"until"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	if input.Until.IsZero() || !input.Until.After(time.Now().UTC()) {
		writeError(w, http.StatusBadRequest, "until must be a future timestamp")
		return
	}
	user, err := a.users.Ban(id, input.Until, input.Reason)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeAPI(w, http.StatusOK, map[string]any{"user": publicUser(user, "password"), "reason": input.Reason, "until": input.Until.UTC()})
}

func (a *app) currentUser(w http.ResponseWriter, r *http.Request) (users.User, bool) {
	authHeader := r.Header.Get("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		writeError(w, http.StatusUnauthorized, "bearer token required")
		return users.User{}, false
	}

	userID, err := a.auth.ParseToken(token)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return users.User{}, false
	}

	user, ok := a.users.FindByID(userID)
	if !ok {
		writeError(w, http.StatusUnauthorized, "user not found")
		return users.User{}, false
	}
	if a.users.IsBanned(user, time.Now().UTC()) {
		writeError(w, http.StatusForbidden, "user is banned: "+user.BanReason)
		return users.User{}, false
	}

	return user, true
}

func (a *app) currentAdmin(w http.ResponseWriter, r *http.Request) (users.User, bool) {
	user, ok := a.currentUser(w, r)
	if !ok {
		return users.User{}, false
	}
	if !admin.CanManage(user) {
		writeError(w, http.StatusForbidden, "admin role required")
		return users.User{}, false
	}
	return user, true
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return false
	}

	return true
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, response{"code": status, "message": message, "data": nil, "success": false, "error": message})
}

func writeAPI(w http.ResponseWriter, status int, data any) {
	writeJSON(w, status, response{"code": 0, "message": "success", "data": data, "success": true})
}

func publicUser(user users.User, method string) map[string]any {
	initial := "?"
	if user.Username != "" {
		initial = strings.ToUpper(string([]rune(user.Username)[0]))
	}
	return map[string]any{"id": user.ID, "name": user.Username, "initial": initial, "avatar": "", "level": "lv1", "role": user.Role, "method": method}
}

func publishedPosts(items []catalog.ForumPost) []catalog.ForumPost {
	result := make([]catalog.ForumPost, 0, len(items))
	for _, item := range items {
		if item.Status == "published" {
			result = append(result, item)
		}
	}
	return result
}

func paginate(items any, r *http.Request) map[string]any {
	page, pageSize := positiveInt(r.URL.Query().Get("page"), 1), positiveInt(r.URL.Query().Get("pageSize"), 20)
	// Keep pagination generic for the phase-one read APIs while preserving typed slices.
	value := reflect.ValueOf(items)
	total := value.Len()
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return map[string]any{"page": page, "pageSize": pageSize, "total": total, "items": value.Slice(start, end).Interface()}
}

func positiveInt(value string, fallback int) int {
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func routeID(w http.ResponseWriter, r *http.Request, resource string) (int64, bool) {
	_ = resource
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

func postIDFromPath(w http.ResponseWriter, r *http.Request) (int64, bool) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	for index, part := range parts {
		if part == "posts" && index+1 < len(parts) {
			id, err := strconv.ParseInt(parts[index+1], 10, 64)
			if err != nil || id <= 0 {
				writeError(w, http.StatusBadRequest, "invalid id")
				return 0, false
			}

			return id, true
		}
	}

	if len(parts) < 1 {
		writeError(w, http.StatusBadRequest, "id is required")
		return 0, false
	}

	writeError(w, http.StatusBadRequest, "id is required")
	return 0, false
}

func sanitizeHeader(value string) string {
	return strings.NewReplacer("\r", "", "\n", "", `"`, "").Replace(value)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
