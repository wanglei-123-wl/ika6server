package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/wanglei-123-wl/ika6server/backend/internal/admin"
	"github.com/wanglei-123-wl/ika6server/backend/internal/audit"
	"github.com/wanglei-123-wl/ika6server/backend/internal/auth"
	"github.com/wanglei-123-wl/ika6server/backend/internal/blocklist"
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
	users      *users.Store
	auth       *auth.Service
	posts      *posts.Store
	files      *files.Store
	audit      *audit.Service
	blocklist  *blocklist.Store
	reputation *reputation.Store
	httpServer *http.Server
}

type response map[string]any

func main() {
	cfg := config.Load()
	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	userStore := users.NewStore()
	postStore := posts.NewStore()
	blocklistStore := blocklist.NewStore()
	reputationStore := reputation.NewStore()
	application := &app{
		config:   cfg,
		database: db,
		users:    userStore,
		auth:     auth.NewService(userStore, cfg.TokenSecret),
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
	mux.HandleFunc("GET /api/users/me", a.handleMe)
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
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}

	user, token, err := a.auth.Register(input.Username, input.Email, input.Password)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	a.reputation.Add(user.ID, reputation.EventRegister, 5, "registered account")

	writeJSON(w, http.StatusCreated, response{
		"ok":    true,
		"user":  user,
		"token": token,
	})
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
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}

	user, token, err := a.auth.Login(input.Email, input.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, response{
		"ok":    true,
		"user":  user,
		"token": token,
	})
}

func (a *app) handleMe(w http.ResponseWriter, r *http.Request) {
	user, ok := a.currentUser(w, r)
	if !ok {
		return
	}

	writeJSON(w, http.StatusOK, response{
		"ok":   true,
		"user": user,
	})
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
	items := search.Posts(a.posts.List(posts.StatusApproved), r.URL.Query().Get("q"))
	writeJSON(w, http.StatusOK, response{
		"ok":    true,
		"items": items,
	})
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
	writeJSON(w, status, response{
		"ok":    false,
		"error": message,
	})
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
