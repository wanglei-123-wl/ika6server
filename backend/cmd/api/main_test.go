package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/wanglei-123-wl/ika6server/backend/internal/auth"
	"github.com/wanglei-123-wl/ika6server/backend/internal/catalog"
	"github.com/wanglei-123-wl/ika6server/backend/internal/config"
	"github.com/wanglei-123-wl/ika6server/backend/internal/database"
	"github.com/wanglei-123-wl/ika6server/backend/internal/posts"
	"github.com/wanglei-123-wl/ika6server/backend/internal/reputation"
	"github.com/wanglei-123-wl/ika6server/backend/internal/users"
)

func testApp() *app {
	userStore := users.NewStoreWithAdmin("admin@test.com")
	return &app{
		config:     config.Config{TokenSecret: "test-secret"},
		database:   &database.Database{},
		users:      userStore,
		auth:       auth.NewService(userStore, "test-secret"),
		posts:      posts.NewStore(),
		catalog:    catalog.NewStore(),
		reputation: reputation.NewStore(),
	}
}

func testHandler(a *app) http.Handler {
	mux := http.NewServeMux()
	a.routes(mux)
	return withCORS(mux)
}

func TestPhaseOneContracts(t *testing.T) {
	a := testApp()
	handler := testHandler(a)

	register := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{"name":"admin","account":"admin@test.com","password":"12345678"}`))
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(register, req)
	if register.Code != http.StatusCreated {
		t.Fatalf("register status = %d, body = %s", register.Code, register.Body.String())
	}

	var registered struct {
		Code int `json:"code"`
		Data struct {
			Token string `json:"token"`
			User  struct {
				Name string `json:"name"`
				Role string `json:"role"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(register.Body.Bytes(), &registered); err != nil {
		t.Fatal(err)
	}
	if registered.Code != 0 || registered.Data.Token == "" || registered.Data.User.Name != "admin" || registered.Data.User.Role != "admin" {
		t.Fatalf("unexpected register response: %s", register.Body.String())
	}

	me := httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+registered.Data.Token)
	handler.ServeHTTP(me, req)
	if me.Code != http.StatusOK {
		t.Fatalf("me status = %d, body = %s", me.Code, me.Body.String())
	}

	for _, path := range []string{"/api/home", "/api/games", "/api/forum/bars", "/api/forum/posts", "/api/repos", "/api/dev-docs"} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path+"?page=1&pageSize=1", nil))
		if recorder.Code != http.StatusOK {
			t.Fatalf("%s status = %d, body = %s", path, recorder.Code, recorder.Body.String())
		}
		var envelope struct {
			Code int            `json:"code"`
			Data map[string]any `json:"data"`
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		if envelope.Code != 0 || envelope.Data == nil {
			t.Fatalf("%s returned invalid envelope: %s", path, recorder.Body.String())
		}
	}
	fileKind := httptest.NewRecorder()
	handler.ServeHTTP(fileKind, httptest.NewRequest(http.MethodGet, "/api/games/1/files/unknown", nil))
	if fileKind.Code != http.StatusBadRequest {
		t.Fatalf("invalid file kind status = %d, body = %s", fileKind.Code, fileKind.Body.String())
	}

	search := httptest.NewRecorder()
	handler.ServeHTTP(search, httptest.NewRequest(http.MethodGet, "/api/search?keyword=void", nil))
	if search.Code != http.StatusOK || !strings.Contains(search.Body.String(), `"games"`) || !strings.Contains(search.Body.String(), `"repos"`) {
		t.Fatalf("unexpected search response: %s", search.Body.String())
	}

	repoDownload := httptest.NewRecorder()
	handler.ServeHTTP(repoDownload, httptest.NewRequest(http.MethodGet, "/api/repos/1/download", nil))
	if repoDownload.Code != http.StatusUnauthorized {
		t.Fatalf("repository download without auth status = %d, body = %s", repoDownload.Code, repoDownload.Body.String())
	}
}

func TestAuthMeRequiresBearerToken(t *testing.T) {
	a := testApp()
	recorder := httptest.NewRecorder()
	testHandler(a).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/auth/me", nil))
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}

func registerTestUser(t *testing.T, handler http.Handler) string {
	t.Helper()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{"name":"builder","account":"builder@test.com","password":"12345678"}`))
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(recorder, request)
	var envelope struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	return envelope.Data.Token
}

func TestPhaseTwoInteractions(t *testing.T) {
	a := testApp()
	handler := testHandler(a)
	token := registerTestUser(t, handler)

	repoDownload := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/repos/1/download", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(repoDownload, request)
	if repoDownload.Code != http.StatusOK || !strings.Contains(repoDownload.Body.String(), `"downloadUrl"`) {
		t.Fatalf("repository download failed: %d %s", repoDownload.Code, repoDownload.Body.String())
	}

	like := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/games/1/like", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(like, request)
	if like.Code != http.StatusOK || !strings.Contains(like.Body.String(), `"changed":true`) {
		t.Fatalf("game like failed: %d %s", like.Code, like.Body.String())
	}

	duplicateLike := httptest.NewRecorder()
	handler.ServeHTTP(duplicateLike, request)
	if duplicateLike.Code != http.StatusOK || !strings.Contains(duplicateLike.Body.String(), `"changed":false`) {
		t.Fatalf("duplicate game like was not idempotent: %d %s", duplicateLike.Code, duplicateLike.Body.String())
	}

	play := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/games/1/play", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(play, request)
	if play.Code != http.StatusOK || !strings.Contains(play.Body.String(), `"plays":"18.4k"`) {
		t.Fatalf("game play failed: %d %s", play.Code, play.Body.String())
	}

	post := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/forum/posts", strings.NewReader(`{"title":"新帖","cat":"交流","tags":["测试"],"content":"正文","barId":1}`))
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(post, request)
	if post.Code != http.StatusCreated || !strings.Contains(post.Body.String(), `"测试"`) {
		t.Fatalf("forum post failed: %d %s", post.Code, post.Body.String())
	}

	postLike := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/forum/posts/2/like", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(postLike, request)
	if postLike.Code != http.StatusOK || !strings.Contains(postLike.Body.String(), `"changed":true`) {
		t.Fatalf("forum like failed: %d %s", postLike.Code, postLike.Body.String())
	}

	reply := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/forum/posts/2/replies", strings.NewReader(`{"content":"支持一下"}`))
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(reply, request)
	if reply.Code != http.StatusCreated || !strings.Contains(reply.Body.String(), `"postId":2`) {
		t.Fatalf("reply failed: %d %s", reply.Code, reply.Body.String())
	}

	game := httptest.NewRecorder()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("title", "新游戏")
	_ = writer.WriteField("summary", "测试游戏")
	_ = writer.WriteField("engine", "Godot 4")
	_ = writer.WriteField("genre", "Puzzle")
	_ = writer.WriteField("license", "MIT")
	_ = writer.Close()
	request = httptest.NewRequest(http.MethodPost, "/api/games", &body)
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	handler.ServeHTTP(game, request)
	if game.Code != http.StatusCreated || !strings.Contains(game.Body.String(), `"status":"reviewing"`) {
		t.Fatalf("game submission failed: %d %s", game.Code, game.Body.String())
	}
}

func TestPhaseThreeAdminControls(t *testing.T) {
	a := testApp()
	handler := testHandler(a)
	adminToken := registerTestUserWithCredentials(t, handler, "admin", "admin@test.com")
	userToken := registerTestUserWithCredentials(t, handler, "member", "member@test.com")

	dashboard := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/admin/dashboard", nil)
	request.Header.Set("Authorization", "Bearer "+adminToken)
	handler.ServeHTTP(dashboard, request)
	if dashboard.Code != http.StatusOK || !strings.Contains(dashboard.Body.String(), `"pendingGames":0`) {
		t.Fatalf("admin dashboard failed: %d %s", dashboard.Code, dashboard.Body.String())
	}

	forbidden := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/admin/dashboard", nil)
	request.Header.Set("Authorization", "Bearer "+userToken)
	handler.ServeHTTP(forbidden, request)
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("regular user admin access status = %d, body = %s", forbidden.Code, forbidden.Body.String())
	}

	game := httptest.NewRecorder()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("title", "待审核游戏")
	_ = writer.WriteField("summary", "待审核")
	_ = writer.Close()
	request = httptest.NewRequest(http.MethodPost, "/api/games", &body)
	request.Header.Set("Authorization", "Bearer "+userToken)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	handler.ServeHTTP(game, request)
	if game.Code != http.StatusCreated {
		t.Fatalf("game submission failed: %d %s", game.Code, game.Body.String())
	}

	review := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/admin/games/3/review", strings.NewReader(`{"status":"approved","reason":"符合要求"}`))
	request.Header.Set("Authorization", "Bearer "+adminToken)
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(review, request)
	if review.Code != http.StatusOK || !strings.Contains(review.Body.String(), `"status":"published"`) {
		t.Fatalf("game review failed: %d %s", review.Code, review.Body.String())
	}

	ban := httptest.NewRecorder()
	until := time.Now().UTC().Add(time.Hour).Format(time.RFC3339)
	request = httptest.NewRequest(http.MethodPost, "/api/admin/users/2/ban", strings.NewReader(`{"reason":"违规发言","until":"`+until+`"}`))
	request.Header.Set("Authorization", "Bearer "+adminToken)
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(ban, request)
	if ban.Code != http.StatusOK {
		t.Fatalf("user ban failed: %d %s", ban.Code, ban.Body.String())
	}

	bannedWrite := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/forum/posts", strings.NewReader(`{"title":"被禁言用户","content":"不能发","barId":1}`))
	request.Header.Set("Authorization", "Bearer "+userToken)
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(bannedWrite, request)
	if bannedWrite.Code != http.StatusForbidden {
		t.Fatalf("banned write status = %d, body = %s", bannedWrite.Code, bannedWrite.Body.String())
	}
}

func TestGameUploadRejectsInvalidFileBeforeScan(t *testing.T) {
	a := testApp()
	handler := testHandler(a)
	token := registerTestUser(t, handler)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("title", "带坏文件的游戏")
	part, err := writer.CreateFormFile("sourceFile", "payload.ps1")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte("Write-Host test"))
	_ = writer.Close()

	request := httptest.NewRequest(http.MethodPost, "/api/games", &body)
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnprocessableEntity || !strings.Contains(recorder.Body.String(), "package file type is not allowed") {
		t.Fatalf("invalid upload status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestDevDocsUpdateRequiresDatabaseMode(t *testing.T) {
	a := testApp()
	handler := testHandler(a)
	token := registerTestUserWithCredentials(t, handler, "admin", "admin@test.com")
	request := httptest.NewRequest(http.MethodPut, "/api/admin/dev-docs", strings.NewReader(`{"title":"更新后的文档","sub":"说明","steps":["一步"],"code":"go test"}`))
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNotImplemented {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}

func registerTestUserWithCredentials(t *testing.T, handler http.Handler, name, account string) string {
	t.Helper()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{"name":"`+name+`","account":"`+account+`","password":"12345678"}`))
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(recorder, request)
	var envelope struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	return envelope.Data.Token
}
