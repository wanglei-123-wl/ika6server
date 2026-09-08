package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/wanglei-123-wl/ika6server/backend/internal/audit"
	"github.com/wanglei-123-wl/ika6server/backend/internal/auth"
	"github.com/wanglei-123-wl/ika6server/backend/internal/catalog"
	"github.com/wanglei-123-wl/ika6server/backend/internal/config"
	"github.com/wanglei-123-wl/ika6server/backend/internal/database"
	playdeploy "github.com/wanglei-123-wl/ika6server/backend/internal/play"
	"github.com/wanglei-123-wl/ika6server/backend/internal/posts"
	"github.com/wanglei-123-wl/ika6server/backend/internal/reputation"
	"github.com/wanglei-123-wl/ika6server/backend/internal/users"
)

func testApp() *app {
	userStore := users.NewStoreWithAdmin("admin@test.com")
	return &app{
		config:      config.Config{TokenSecret: "test-secret"},
		database:    &database.Database{},
		users:       userStore,
		auth:        auth.NewService(userStore, "test-secret"),
		posts:       posts.NewStore(),
		play:        playdeploy.NewService(""),
		auditLog:    audit.NewMemoryLogger(),
		catalog:     catalog.NewStore(),
		reputation:  reputation.NewStore(),
		reviewUrges: make(map[string]time.Time),
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
	if !strings.Contains(recorder.Body.String(), `"errorCode":"INVALID_TOKEN"`) {
		t.Fatalf("missing token error code missing: %s", recorder.Body.String())
	}
}

func TestHealthReportsPersistenceMode(t *testing.T) {
	a := testApp()
	recorder := httptest.NewRecorder()
	testHandler(a).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("health status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"persistence":"memory"`) {
		t.Fatalf("health response missing memory persistence mode: %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"databaseConfigured":false`) {
		t.Fatalf("health response missing databaseConfigured=false: %s", recorder.Body.String())
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	a := testApp()
	handler := testHandler(a)
	_ = registerTestUserWithCredentials(t, handler, "builder", "builder@test.com")

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"account":"builder@test.com","password":"wrong-password"}`))
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("wrong password login status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"errorCode":"INVALID_CREDENTIALS"`) {
		t.Fatalf("wrong password error code missing: %s", recorder.Body.String())
	}
}

func TestLoginRejectsWrongAccountWithKnownPassword(t *testing.T) {
	a := testApp()
	handler := testHandler(a)
	_ = registerTestUserWithCredentials(t, handler, "builder", "builder@test.com")

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"account":"unknown@test.com","password":"12345678"}`))
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("wrong account login status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"errorCode":"INVALID_CREDENTIALS"`) {
		t.Fatalf("wrong account error code missing: %s", recorder.Body.String())
	}
}

func TestRegisterRejectsDuplicateEmailWithStableErrorCode(t *testing.T) {
	a := testApp()
	handler := testHandler(a)
	_ = registerTestUserWithCredentials(t, handler, "builder", "builder@test.com")

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{"name":"another","account":"builder@test.com","password":"12345678"}`))
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusConflict {
		t.Fatalf("duplicate email status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"errorCode":"EMAIL_EXISTS"`) {
		t.Fatalf("duplicate email error code missing: %s", recorder.Body.String())
	}
}

func TestLogoutRevokesCurrentToken(t *testing.T) {
	a := testApp()
	handler := testHandler(a)
	token := registerTestUserWithCredentials(t, handler, "builder", "builder@test.com")

	logout := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(logout, request)
	if logout.Code != http.StatusOK {
		t.Fatalf("logout status = %d, body = %s", logout.Code, logout.Body.String())
	}

	me := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(me, request)
	if me.Code != http.StatusUnauthorized {
		t.Fatalf("revoked token status = %d, body = %s", me.Code, me.Body.String())
	}

	login := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"account":"builder@test.com","password":"12345678"}`))
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(login, request)
	if login.Code != http.StatusOK {
		t.Fatalf("fresh login status = %d, body = %s", login.Code, login.Body.String())
	}
	var envelope struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(login.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.Token == "" || envelope.Data.Token == token {
		t.Fatalf("expected fresh non-empty token, got %q", envelope.Data.Token)
	}
}

func TestSocialLoginCreatesAndReusesLocalAccount(t *testing.T) {
	a := testApp()
	handler := testHandler(a)

	first := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/auth/social-login", strings.NewReader(`{"provider":"GitHub","account":"octo@test.com","name":"octo"}`))
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(first, request)
	if first.Code != http.StatusOK {
		t.Fatalf("social login status = %d, body = %s", first.Code, first.Body.String())
	}
	var firstEnvelope struct {
		Data struct {
			Token string `json:"token"`
			User  struct {
				ID     int64  `json:"id"`
				Method string `json:"method"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(first.Body.Bytes(), &firstEnvelope); err != nil {
		t.Fatal(err)
	}
	if firstEnvelope.Data.Token == "" || firstEnvelope.Data.User.Method != "GitHub" {
		t.Fatalf("unexpected social login response: %s", first.Body.String())
	}

	second := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/auth/social-login", strings.NewReader(`{"method":"GitHub","email":"octo@test.com","username":"ignored"}`))
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(second, request)
	if second.Code != http.StatusOK {
		t.Fatalf("repeat social login status = %d, body = %s", second.Code, second.Body.String())
	}
	var secondEnvelope struct {
		Data struct {
			Token string `json:"token"`
			User  struct {
				ID int64 `json:"id"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(second.Body.Bytes(), &secondEnvelope); err != nil {
		t.Fatal(err)
	}
	if secondEnvelope.Data.Token == "" || secondEnvelope.Data.User.ID != firstEnvelope.Data.User.ID {
		t.Fatalf("expected social login to reuse user, first=%s second=%s", first.Body.String(), second.Body.String())
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

	replyLike := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/forum/replies/1/like", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(replyLike, request)
	if replyLike.Code != http.StatusOK || !strings.Contains(replyLike.Body.String(), `"likes":1`) || !strings.Contains(replyLike.Body.String(), `"changed":true`) {
		t.Fatalf("reply like failed: %d %s", replyLike.Code, replyLike.Body.String())
	}

	duplicateReplyLike := httptest.NewRecorder()
	handler.ServeHTTP(duplicateReplyLike, request)
	if duplicateReplyLike.Code != http.StatusOK || !strings.Contains(duplicateReplyLike.Body.String(), `"likes":1`) || !strings.Contains(duplicateReplyLike.Body.String(), `"changed":false`) {
		t.Fatalf("duplicate reply like was not idempotent: %d %s", duplicateReplyLike.Code, duplicateReplyLike.Body.String())
	}

	replyList := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/forum/posts/2/replies?page=1&pageSize=20", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(replyList, request)
	if replyList.Code != http.StatusOK || !strings.Contains(replyList.Body.String(), `"liked":true`) {
		t.Fatalf("reply list did not include current user like state: %d %s", replyList.Code, replyList.Body.String())
	}

	comment := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/forum/posts/2/comments", strings.NewReader(`{"content":"一级评论"}`))
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(comment, request)
	if comment.Code != http.StatusCreated || !strings.Contains(comment.Body.String(), `"parentId":null`) {
		t.Fatalf("comment failed: %d %s", comment.Code, comment.Body.String())
	}
	var commentEnvelope struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(comment.Body.Bytes(), &commentEnvelope); err != nil {
		t.Fatal(err)
	}

	commentReply := httptest.NewRecorder()
	commentID := catalog.IDString(commentEnvelope.Data.ID)
	request = httptest.NewRequest(http.MethodPost, "/api/forum/posts/2/comments", strings.NewReader(`{"content":"二级回复","parentId":`+commentID+`,"replyToCommentId":`+commentID+`,"replyToAuthor":"builder"}`))
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(commentReply, request)
	if commentReply.Code != http.StatusCreated || !strings.Contains(commentReply.Body.String(), `"parentId":`+commentID) || !strings.Contains(commentReply.Body.String(), `"replyToAuthor":"builder"`) {
		t.Fatalf("comment reply failed: %d %s", commentReply.Code, commentReply.Body.String())
	}
	var commentReplyEnvelope struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(commentReply.Body.Bytes(), &commentReplyEnvelope); err != nil {
		t.Fatal(err)
	}

	commentReplyLike := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/forum/comments/"+catalog.IDString(commentReplyEnvelope.Data.ID)+"/like", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(commentReplyLike, request)
	if commentReplyLike.Code != http.StatusOK || !strings.Contains(commentReplyLike.Body.String(), `"likes":1`) || !strings.Contains(commentReplyLike.Body.String(), `"changed":true`) {
		t.Fatalf("comment reply like failed: %d %s", commentReplyLike.Code, commentReplyLike.Body.String())
	}

	commentReplies := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/forum/comments/"+catalog.IDString(commentEnvelope.Data.ID)+"/replies?page=1&pageSize=20", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(commentReplies, request)
	if commentReplies.Code != http.StatusOK || !strings.Contains(commentReplies.Body.String(), `"liked":true`) || !strings.Contains(commentReplies.Body.String(), `"replyToCommentId":`+commentID) {
		t.Fatalf("comment replies failed: %d %s", commentReplies.Code, commentReplies.Body.String())
	}

	comments := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/forum/posts/2/comments?page=1&pageSize=20", nil)
	handler.ServeHTTP(comments, request)
	if comments.Code != http.StatusOK || !strings.Contains(comments.Body.String(), `"replyCount":1`) {
		t.Fatalf("comments list failed: %d %s", comments.Code, comments.Body.String())
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

func TestThreadedCommentGuards(t *testing.T) {
	a := testApp()
	handler := testHandler(a)
	token := registerTestUserWithCredentials(t, handler, "builder", "builder@test.com")

	unauthorizedCreate := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/forum/posts/1/comments", strings.NewReader(`{"content":"未登录评论"}`))
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(unauthorizedCreate, request)
	if unauthorizedCreate.Code != http.StatusUnauthorized || !strings.Contains(unauthorizedCreate.Body.String(), `"errorCode":"INVALID_TOKEN"`) {
		t.Fatalf("unauthorized comment status = %d, body = %s", unauthorizedCreate.Code, unauthorizedCreate.Body.String())
	}

	parentID := createThreadedComment(t, handler, token, 1, `{"content":"一级评论"}`)
	childID := createThreadedComment(t, handler, token, 1, `{"content":"二级回复","parentId":`+catalog.IDString(parentID)+`,"replyToCommentId":`+catalog.IDString(parentID)+`}`)

	grandchild := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/forum/posts/1/comments", strings.NewReader(`{"content":"三级回复","parentId":`+catalog.IDString(childID)+`,"replyToCommentId":`+catalog.IDString(childID)+`}`))
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(grandchild, request)
	if grandchild.Code != http.StatusBadRequest {
		t.Fatalf("grandchild comment status = %d, body = %s", grandchild.Code, grandchild.Body.String())
	}

	wrongTarget := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/forum/posts/1/comments", strings.NewReader(`{"content":"错误目标","parentId":`+catalog.IDString(parentID)+`,"replyToCommentId":99999}`))
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(wrongTarget, request)
	if wrongTarget.Code != http.StatusBadRequest {
		t.Fatalf("wrong reply target status = %d, body = %s", wrongTarget.Code, wrongTarget.Body.String())
	}

	childReplies := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/forum/comments/"+catalog.IDString(childID)+"/replies", nil)
	handler.ServeHTTP(childReplies, request)
	if childReplies.Code != http.StatusNotFound {
		t.Fatalf("child replies status = %d, body = %s", childReplies.Code, childReplies.Body.String())
	}

	unauthorizedLike := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/forum/comments/"+catalog.IDString(parentID)+"/like", nil)
	handler.ServeHTTP(unauthorizedLike, request)
	if unauthorizedLike.Code != http.StatusUnauthorized || !strings.Contains(unauthorizedLike.Body.String(), `"errorCode":"INVALID_TOKEN"`) {
		t.Fatalf("unauthorized comment like status = %d, body = %s", unauthorizedLike.Code, unauthorizedLike.Body.String())
	}

	invalidTokenRead := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/forum/posts/1/comments", nil)
	request.Header.Set("Authorization", "Bearer invalid-token")
	handler.ServeHTTP(invalidTokenRead, request)
	if invalidTokenRead.Code != http.StatusUnauthorized || !strings.Contains(invalidTokenRead.Body.String(), `"errorCode":"INVALID_TOKEN"`) {
		t.Fatalf("invalid token read status = %d, body = %s", invalidTokenRead.Code, invalidTokenRead.Body.String())
	}
}

func TestDeveloperCenterContracts(t *testing.T) {
	a := testApp()
	handler := testHandler(a)
	token := registerTestUserWithCredentials(t, handler, "builder", "builder@test.com")
	otherToken := registerTestUserWithCredentials(t, handler, "other", "other@test.com")

	reviewing, err := a.catalog.AddGame("@builder", "动物混战", "summary", "HTML5", "休闲", "MIT")
	if err != nil {
		t.Fatal(err)
	}
	rejected, err := a.catalog.AddGame("@builder", "被退回作品", "summary", "Godot", "动作", "MIT")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.catalog.ReviewGame(rejected.ID, "rejected"); err != nil {
		t.Fatal(err)
	}
	otherGame, err := a.catalog.AddGame("@other", "别人的作品", "summary", "Unity", "冒险", "MIT")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.catalog.ReviewGame(otherGame.ID, "rejected"); err != nil {
		t.Fatal(err)
	}

	me := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/developer/me", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(me, request)
	if me.Code != http.StatusOK || !strings.Contains(me.Body.String(), `"name":"builder"`) || !strings.Contains(me.Body.String(), `"engines":[]`) {
		t.Fatalf("developer me failed: %d %s", me.Code, me.Body.String())
	}

	stats := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/developer/stats", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(stats, request)
	if stats.Code != http.StatusOK || !strings.Contains(stats.Body.String(), `"works":2`) {
		t.Fatalf("developer stats failed: %d %s", stats.Code, stats.Body.String())
	}

	games := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/developer/games?status=all", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(games, request)
	if games.Code != http.StatusOK || !strings.Contains(games.Body.String(), `"title":"动物混战"`) || strings.Contains(games.Body.String(), `"别人的作品"`) {
		t.Fatalf("developer games failed: %d %s", games.Code, games.Body.String())
	}

	urge := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/developer/games/"+catalog.IDString(reviewing.ID)+"/urge-review", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(urge, request)
	if urge.Code != http.StatusOK || !strings.Contains(urge.Body.String(), `"changed":true`) {
		t.Fatalf("developer urge failed: %d %s", urge.Code, urge.Body.String())
	}
	duplicateUrge := httptest.NewRecorder()
	handler.ServeHTTP(duplicateUrge, request)
	if duplicateUrge.Code != http.StatusOK || !strings.Contains(duplicateUrge.Body.String(), `"changed":false`) {
		t.Fatalf("duplicate developer urge failed: %d %s", duplicateUrge.Code, duplicateUrge.Body.String())
	}

	forbiddenResubmit := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/developer/games/"+catalog.IDString(rejected.ID)+"/resubmit", nil)
	request.Header.Set("Authorization", "Bearer "+otherToken)
	handler.ServeHTTP(forbiddenResubmit, request)
	if forbiddenResubmit.Code != http.StatusForbidden {
		t.Fatalf("foreign resubmit status = %d, body = %s", forbiddenResubmit.Code, forbiddenResubmit.Body.String())
	}

	resubmit := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/developer/games/"+catalog.IDString(rejected.ID)+"/resubmit", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(resubmit, request)
	if resubmit.Code != http.StatusOK || !strings.Contains(resubmit.Body.String(), `"status":"reviewing"`) {
		t.Fatalf("developer resubmit failed: %d %s", resubmit.Code, resubmit.Body.String())
	}

	deletePublished := httptest.NewRecorder()
	if _, err := a.catalog.ReviewGame(rejected.ID, "approved"); err != nil {
		t.Fatal(err)
	}
	request = httptest.NewRequest(http.MethodDelete, "/api/developer/games/"+catalog.IDString(rejected.ID), nil)
	request.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(deletePublished, request)
	if deletePublished.Code != http.StatusConflict {
		t.Fatalf("published delete status = %d, body = %s", deletePublished.Code, deletePublished.Body.String())
	}

	deletable, err := a.catalog.AddGame("@builder", "可删除作品", "summary", "HTML5", "休闲", "MIT")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.catalog.ReviewGame(deletable.ID, "rejected"); err != nil {
		t.Fatal(err)
	}
	deleteGame := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodDelete, "/api/developer/games/"+catalog.IDString(deletable.ID), nil)
	request.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(deleteGame, request)
	if deleteGame.Code != http.StatusOK || !strings.Contains(deleteGame.Body.String(), `"message":"已删除"`) {
		t.Fatalf("developer delete failed: %d %s", deleteGame.Code, deleteGame.Body.String())
	}
}

func TestPlayAssetRequiresPublishedGame(t *testing.T) {
	a := testApp()
	a.config.PlayDir = t.TempDir()
	a.play = playdeploy.NewService(a.config.PlayDir)
	handler := testHandler(a)
	token := registerTestUserWithCredentials(t, handler, "player", "player@test.com")
	game, err := a.catalog.AddGame("@builder", "试玩游戏", "summary", "Phaser", "Arcade", "MIT")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(a.config.PlayDir, catalog.IDString(game.ID)), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(a.config.PlayDir, catalog.IDString(game.ID), "index.html"), []byte("<h1>play</h1>"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := a.catalog.SetGamePlayURL(game.ID, "/play/games/"+catalog.IDString(game.ID)+"/index.html"); err != nil {
		t.Fatal(err)
	}

	beforeReview := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/play/games/"+catalog.IDString(game.ID)+"/index.html", nil)
	handler.ServeHTTP(beforeReview, request)
	if beforeReview.Code != http.StatusNotFound {
		t.Fatalf("unpublished play asset status = %d, body = %s", beforeReview.Code, beforeReview.Body.String())
	}

	playBeforeReview := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/games/"+catalog.IDString(game.ID)+"/play", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(playBeforeReview, request)
	if playBeforeReview.Code != http.StatusNotFound {
		t.Fatalf("unpublished play status = %d, body = %s", playBeforeReview.Code, playBeforeReview.Body.String())
	}

	if _, err := a.catalog.ReviewGame(game.ID, "approved"); err != nil {
		t.Fatal(err)
	}
	afterReview := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/play/games/"+catalog.IDString(game.ID)+"/index.html", nil)
	handler.ServeHTTP(afterReview, request)
	if afterReview.Code != http.StatusOK || !strings.Contains(afterReview.Body.String(), "play") {
		t.Fatalf("published play asset status = %d, body = %s", afterReview.Code, afterReview.Body.String())
	}

	playAfterReview := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/games/"+catalog.IDString(game.ID)+"/play", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(playAfterReview, request)
	if playAfterReview.Code != http.StatusOK || !strings.Contains(playAfterReview.Body.String(), `"playUrl":"/play/games/`) {
		t.Fatalf("published play status = %d, body = %s", playAfterReview.Code, playAfterReview.Body.String())
	}
}

func TestGameSourceRequiresPublishedGame(t *testing.T) {
	a := testApp()
	handler := testHandler(a)
	token := registerTestUserWithCredentials(t, handler, "player", "player@test.com")
	game, err := a.catalog.AddGame("@builder", "源码游戏", "summary", "Godot", "Puzzle", "MIT")
	if err != nil {
		t.Fatal(err)
	}
	if err := a.catalog.SetGameFileURL(game.ID, "source", "/api/games/"+catalog.IDString(game.ID)+"/files/source"); err != nil {
		t.Fatal(err)
	}

	beforeReview := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/games/"+catalog.IDString(game.ID)+"/download-source", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(beforeReview, request)
	if beforeReview.Code != http.StatusNotFound {
		t.Fatalf("unpublished source status = %d, body = %s", beforeReview.Code, beforeReview.Body.String())
	}

	if _, err := a.catalog.ReviewGame(game.ID, "approved"); err != nil {
		t.Fatal(err)
	}
	afterReview := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/games/"+catalog.IDString(game.ID)+"/download-source", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(afterReview, request)
	if afterReview.Code != http.StatusOK || !strings.Contains(afterReview.Body.String(), `"downloadUrl"`) {
		t.Fatalf("published source status = %d, body = %s", afterReview.Code, afterReview.Body.String())
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

	auditLogs := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/admin/audit-logs", nil)
	request.Header.Set("Authorization", "Bearer "+adminToken)
	handler.ServeHTTP(auditLogs, request)
	if auditLogs.Code != http.StatusOK || !strings.Contains(auditLogs.Body.String(), `"audit_logs"`) && !strings.Contains(auditLogs.Body.String(), `"review_game"`) {
		t.Fatalf("audit logs failed: %d %s", auditLogs.Code, auditLogs.Body.String())
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

func createThreadedComment(t *testing.T, handler http.Handler, token string, postID int64, body string) int64 {
	t.Helper()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/forum/posts/"+catalog.IDString(postID)+"/comments", strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("create threaded comment status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	var envelope struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	return envelope.Data.ID
}
