package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
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
	"github.com/wanglei-123-wl/ika6server/backend/internal/reports"
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
		reports:     reports.NewStore(),
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

func TestConfiguredAdminAccessContract(t *testing.T) {
	a := testApp()
	a.config.AdminAccount = "admin@test.com"
	handler := testHandler(a)

	adminToken := registerTestUserWithCredentials(t, handler, "admin", "admin@test.com")
	me := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	request.Header.Set("Authorization", "Bearer "+adminToken)
	handler.ServeHTTP(me, request)
	if me.Code != http.StatusOK || !strings.Contains(me.Body.String(), `"role":"admin"`) || !strings.Contains(me.Body.String(), `"adminAccess":true`) {
		t.Fatalf("configured admin me failed: %d %s", me.Code, me.Body.String())
	}

	guarded := testApp()
	guarded.config.AdminAccount = "unique-admin@test.com"
	guardedHandler := testHandler(guarded)
	fakeAdminToken := registerTestUserWithCredentials(t, guardedHandler, "admin", "admin@test.com")

	forbidden := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/admin/dashboard", nil)
	request.Header.Set("Authorization", "Bearer "+fakeAdminToken)
	guardedHandler.ServeHTTP(forbidden, request)
	if forbidden.Code != http.StatusForbidden || !strings.Contains(forbidden.Body.String(), `"errorCode":"FORBIDDEN"`) {
		t.Fatalf("misconfigured admin status = %d, body = %s", forbidden.Code, forbidden.Body.String())
	}

	fakeMe := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	request.Header.Set("Authorization", "Bearer "+fakeAdminToken)
	guardedHandler.ServeHTTP(fakeMe, request)
	if fakeMe.Code != http.StatusOK || !strings.Contains(fakeMe.Body.String(), `"role":"user"`) || !strings.Contains(fakeMe.Body.String(), `"adminAccess":false`) {
		t.Fatalf("misconfigured admin public user leaked access: %d %s", fakeMe.Code, fakeMe.Body.String())
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

func TestSocialLoginFailsClosed(t *testing.T) {
	a := testApp()
	handler := testHandler(a)
	_ = registerTestUserWithCredentials(t, handler, "admin", "admin@test.com")
	before, _ := a.users.Counts(time.Now())
	for _, body := range []string{
		`{"provider":"GitHub","account":"new@test.com","name":"new"}`,
		`{"provider":"Google","email":"admin@test.com","username":"admin"}`,
		`{"method":"GitHub","account":"admin@test.com"}`,
	} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/auth/social-login", strings.NewReader(body))
		handler.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusNotImplemented || !strings.Contains(recorder.Body.String(), `"errorCode":"NOT_IMPLEMENTED"`) {
			t.Fatalf("social login must fail closed: %d %s", recorder.Code, recorder.Body.String())
		}
		if strings.Contains(recorder.Body.String(), `"token"`) {
			t.Fatal("unverified social login must not issue a token")
		}
	}
	after, _ := a.users.Counts(time.Now())
	if after != before {
		t.Fatal("disabled social login must not create accounts")
	}
}

type loginLookupFailure struct {
	users.Repository
	err error
}

func (r loginLookupFailure) FindByEmail(string) (users.User, error) {
	return users.User{}, r.err
}

type loginSQLStateError struct{ state string }

func (e loginSQLStateError) Error() string    { return "private-database-detail" }
func (e loginSQLStateError) SQLState() string { return e.state }

func TestLoginDiagnosticsAndErrorClassification(t *testing.T) {
	const password = "private-test-password"
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name      string
		account   string
		password  string
		hash      string
		lookupErr error
		status    int
		errorCode string
		outcome   string
		browser   bool
		sqlState  string
	}{
		{name: "success", account: "admin@test.com", password: password, hash: hash, status: 200, outcome: "success", browser: true},
		{name: "wrong password", account: "admin@test.com", password: "wrong-password", hash: hash, status: 401, errorCode: errorCodeInvalidCredentials, outcome: "password_mismatch"},
		{name: "missing account", account: "missing@test.com", password: password, hash: hash, status: 401, errorCode: errorCodeInvalidCredentials, outcome: "account_not_found"},
		{name: "bad stored hash", account: "admin@test.com", password: password, hash: "private-malformed-hash", status: 503, errorCode: errorCodeAuthUnavailable, outcome: "invalid_password_hash", browser: true},
		{name: "lookup error", account: "admin@test.com", password: password, hash: hash, lookupErr: errors.New("private-database-detail"), status: 503, errorCode: errorCodeAuthUnavailable, outcome: "lookup_error", browser: true},
		{name: "sqlstate", account: "admin@test.com", password: password, hash: hash, lookupErr: loginSQLStateError{"08006"}, status: 503, errorCode: errorCodeAuthUnavailable, outcome: "lookup_error", sqlState: "08006"},
		{name: "unsafe sqlstate", account: "admin@test.com", password: password, hash: hash, lookupErr: loginSQLStateError{"leak\n"}, status: 503, errorCode: errorCodeAuthUnavailable, outcome: "lookup_error"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := testApp()
			a.config.AdminAccount = "admin@test.com"
			if _, err := a.users.Create("admin", "admin@test.com", tc.hash); err != nil {
				t.Fatal(err)
			}
			if tc.lookupErr != nil {
				a.auth = auth.NewService(loginLookupFailure{a.users, tc.lookupErr}, "test-secret")
			}
			var logs bytes.Buffer
			previousOutput := log.Writer()
			log.SetOutput(&logs)
			t.Cleanup(func() { log.SetOutput(previousOutput) })

			body, err := json.Marshal(map[string]any{"account": tc.account, "password": tc.password, "remember": false})
			if err != nil {
				t.Fatal(err)
			}
			request := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
			request.Header.Set("X-Request-ID", "untrusted-client-id")
			if tc.browser {
				request.Header.Set("Origin", "http://browser.test")
			}
			recorder := httptest.NewRecorder()
			testHandler(a).ServeHTTP(recorder, request)
			if recorder.Code != tc.status {
				t.Fatalf("status=%d want=%d body=%s", recorder.Code, tc.status, recorder.Body.String())
			}
			requestID := recorder.Header().Get("X-Request-ID")
			if requestID == "" || requestID == "untrusted-client-id" || !strings.Contains(logs.String(), "request_id="+requestID) {
				t.Fatal("expected a server-generated request ID shared with the log")
			}
			if recorder.Header().Get("Access-Control-Expose-Headers") != "X-Request-ID" {
				t.Fatal("browser must be able to read the request ID")
			}
			if !strings.Contains(logs.String(), "outcome="+tc.outcome) || !strings.Contains(logs.String(), `sqlstate="`+tc.sqlState+`"`) {
				t.Fatalf("wrong diagnostic log: %s", logs.String())
			}
			var payload struct {
				ErrorCode string         `json:"errorCode"`
				RequestID string         `json:"requestId"`
				Data      map[string]any `json:"data"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
				t.Fatal(err)
			}
			if tc.status != http.StatusOK && (payload.ErrorCode != tc.errorCode || payload.RequestID != requestID || payload.Data != nil) {
				t.Fatalf("unexpected error contract: %s", recorder.Body.String())
			}
			if tc.status == http.StatusOK {
				token, ok := payload.Data["token"].(string)
				if !ok || token == "" || strings.Contains(logs.String(), token) {
					t.Fatal("token must be returned only to the client, never logged")
				}
			}
			for _, sensitive := range []string{password, tc.hash, "private-database-detail", tc.account} {
				if strings.Contains(logs.String(), sensitive) || strings.Contains(recorder.Body.String(), sensitive) {
					t.Fatal("credentials or internal details leaked")
				}
			}
		})
	}
}

func TestMalformedLoginHasRequestID(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader("{"))
	testHandler(testApp()).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest || recorder.Header().Get("X-Request-ID") == "" ||
		!strings.Contains(recorder.Body.String(), `"requestId":`) {
		t.Fatalf("malformed login missing diagnostic ID: %d %s", recorder.Code, recorder.Body.String())
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

	game, err := a.catalog.AddGame("@builder", "新游戏", "测试游戏", "Godot 4", "Puzzle", "MIT")
	if err != nil || game.Status != "reviewing" {
		t.Fatalf("game fixture creation failed: %+v, %v", game, err)
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
	if me.Code != http.StatusOK || !strings.Contains(me.Body.String(), `"name":"builder"`) || !strings.Contains(me.Body.String(), `"HTML5"`) || !strings.Contains(me.Body.String(), `"Godot"`) {
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

	publicDetailBeforeReview := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/games/"+catalog.IDString(game.ID), nil)
	handler.ServeHTTP(publicDetailBeforeReview, request)
	if publicDetailBeforeReview.Code != http.StatusNotFound {
		t.Fatalf("unpublished public detail status = %d, body = %s", publicDetailBeforeReview.Code, publicDetailBeforeReview.Body.String())
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

	publicDetailAfterReview := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/games/"+catalog.IDString(game.ID), nil)
	handler.ServeHTTP(publicDetailAfterReview, request)
	if publicDetailAfterReview.Code != http.StatusOK || !strings.Contains(publicDetailAfterReview.Body.String(), `"title":"试玩游戏"`) {
		t.Fatalf("published public detail status = %d, body = %s", publicDetailAfterReview.Code, publicDetailAfterReview.Body.String())
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

	game, err := a.catalog.AddGame("@member", "待审核游戏", "待审核", "HTML5", "休闲", "MIT")
	if err != nil {
		t.Fatal(err)
	}

	adminGames := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/admin/games?status=reviewing", nil)
	request.Header.Set("Authorization", "Bearer "+adminToken)
	handler.ServeHTTP(adminGames, request)
	if adminGames.Code != http.StatusOK || !strings.Contains(adminGames.Body.String(), `"title":"待审核游戏"`) || !strings.Contains(adminGames.Body.String(), `"author":"member"`) {
		t.Fatalf("admin games failed: %d %s", adminGames.Code, adminGames.Body.String())
	}

	review := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/admin/games/"+catalog.IDString(game.ID)+"/review", strings.NewReader(`{"status":"rejected","reason":"源码不完整"}`))
	request.Header.Set("Authorization", "Bearer "+adminToken)
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(review, request)
	if review.Code != http.StatusOK || !strings.Contains(review.Body.String(), `"status":"rejected"`) || !strings.Contains(review.Body.String(), `"reason":"源码不完整"`) {
		t.Fatalf("game rejection failed: %d %s", review.Code, review.Body.String())
	}

	developerGames := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/developer/games?status=rejected", nil)
	request.Header.Set("Authorization", "Bearer "+userToken)
	handler.ServeHTTP(developerGames, request)
	if developerGames.Code != http.StatusOK || !strings.Contains(developerGames.Body.String(), `"rejectReason":"源码不完整"`) {
		t.Fatalf("developer games missing reject reason: %d %s", developerGames.Code, developerGames.Body.String())
	}

	approve := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/admin/games/"+catalog.IDString(game.ID)+"/review", strings.NewReader(`{"status":"approved","reason":"符合要求"}`))
	request.Header.Set("Authorization", "Bearer "+adminToken)
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(approve, request)
	if approve.Code != http.StatusOK || !strings.Contains(approve.Body.String(), `"status":"published"`) {
		t.Fatalf("game approval failed: %d %s", approve.Code, approve.Body.String())
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

	adminUsers := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/admin/users?page=1&pageSize=20", nil)
	request.Header.Set("Authorization", "Bearer "+adminToken)
	handler.ServeHTTP(adminUsers, request)
	if adminUsers.Code != http.StatusOK || !strings.Contains(adminUsers.Body.String(), `"email":"member@test.com"`) || !strings.Contains(adminUsers.Body.String(), `"status":"banned"`) {
		t.Fatalf("admin users failed: %d %s", adminUsers.Code, adminUsers.Body.String())
	}

	bannedWrite := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/forum/posts", strings.NewReader(`{"title":"被禁言用户","content":"不能发","barId":1}`))
	request.Header.Set("Authorization", "Bearer "+userToken)
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(bannedWrite, request)
	if bannedWrite.Code != http.StatusForbidden {
		t.Fatalf("banned write status = %d, body = %s", bannedWrite.Code, bannedWrite.Body.String())
	}

	unban := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/admin/users/2/unban", nil)
	request.Header.Set("Authorization", "Bearer "+adminToken)
	handler.ServeHTTP(unban, request)
	if unban.Code != http.StatusOK || !strings.Contains(unban.Body.String(), `"message":"已解除封禁"`) {
		t.Fatalf("user unban failed: %d %s", unban.Code, unban.Body.String())
	}

	auditLogs := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/admin/audit-logs", nil)
	request.Header.Set("Authorization", "Bearer "+adminToken)
	handler.ServeHTTP(auditLogs, request)
	if auditLogs.Code != http.StatusOK || !strings.Contains(auditLogs.Body.String(), `"audit_logs"`) && !strings.Contains(auditLogs.Body.String(), `"review_game"`) {
		t.Fatalf("audit logs failed: %d %s", auditLogs.Code, auditLogs.Body.String())
	}

	reports := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/admin/reports?page=1&pageSize=20", nil)
	request.Header.Set("Authorization", "Bearer "+adminToken)
	handler.ServeHTTP(reports, request)
	if reports.Code != http.StatusOK || !strings.Contains(reports.Body.String(), `"items":[]`) {
		t.Fatalf("admin reports placeholder failed: %d %s", reports.Code, reports.Body.String())
	}

	resolveReport := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/admin/reports/1/resolve", nil)
	request.Header.Set("Authorization", "Bearer "+adminToken)
	handler.ServeHTTP(resolveReport, request)
	if resolveReport.Code != http.StatusNotFound {
		t.Fatalf("resolve missing report status = %d, body = %s", resolveReport.Code, resolveReport.Body.String())
	}
}

func TestBackendCompletionContracts(t *testing.T) {
	a := testApp()
	handler := testHandler(a)
	adminToken := registerTestUserWithCredentials(t, handler, "admin", "admin@test.com")
	memberToken := registerTestUserWithCredentials(t, handler, "member", "member@test.com")

	createPost := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/forum/posts", strings.NewReader(`{"title":"待审核帖子","content":"请审核","barId":1}`))
	request.Header.Set("Authorization", "Bearer "+memberToken)
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(createPost, request)
	if createPost.Code != http.StatusCreated {
		t.Fatalf("create pending post status = %d, body = %s", createPost.Code, createPost.Body.String())
	}

	adminPosts := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/admin/posts?status=pending&page=1&pageSize=20", nil)
	request.Header.Set("Authorization", "Bearer "+adminToken)
	handler.ServeHTTP(adminPosts, request)
	if adminPosts.Code != http.StatusOK || !strings.Contains(adminPosts.Body.String(), `"title":"待审核帖子"`) {
		t.Fatalf("admin pending posts failed: %d %s", adminPosts.Code, adminPosts.Body.String())
	}

	memberAdminPosts := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/admin/posts?status=pending", nil)
	request.Header.Set("Authorization", "Bearer "+memberToken)
	handler.ServeHTTP(memberAdminPosts, request)
	if memberAdminPosts.Code != http.StatusForbidden {
		t.Fatalf("member admin posts status = %d, body = %s", memberAdminPosts.Code, memberAdminPosts.Body.String())
	}

	createReport := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/forum/reports", strings.NewReader(`{"targetType":"post","targetId":1,"reason":"spam","details":"合同测试"}`))
	request.Header.Set("Authorization", "Bearer "+memberToken)
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(createReport, request)
	if createReport.Code != http.StatusCreated || !strings.Contains(createReport.Body.String(), `"status":"pending"`) {
		t.Fatalf("create report failed: %d %s", createReport.Code, createReport.Body.String())
	}

	dashboard := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/admin/dashboard", nil)
	request.Header.Set("Authorization", "Bearer "+adminToken)
	handler.ServeHTTP(dashboard, request)
	if dashboard.Code != http.StatusOK || !strings.Contains(dashboard.Body.String(), `"reports":1`) {
		t.Fatalf("report dashboard count failed: %d %s", dashboard.Code, dashboard.Body.String())
	}

	reportList := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/admin/reports?status=pending", nil)
	request.Header.Set("Authorization", "Bearer "+adminToken)
	handler.ServeHTTP(reportList, request)
	if reportList.Code != http.StatusOK || !strings.Contains(reportList.Body.String(), `"reason":"spam"`) {
		t.Fatalf("admin report list failed: %d %s", reportList.Code, reportList.Body.String())
	}

	resolveReport := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/admin/reports/1/resolve", strings.NewReader(`{"status":"resolved","resolution":"已核查"}`))
	request.Header.Set("Authorization", "Bearer "+adminToken)
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(resolveReport, request)
	if resolveReport.Code != http.StatusOK || !strings.Contains(resolveReport.Body.String(), `"status":"resolved"`) {
		t.Fatalf("resolve report failed: %d %s", resolveReport.Code, resolveReport.Body.String())
	}

	resolveAgain := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/admin/reports/1/resolve", nil)
	request.Header.Set("Authorization", "Bearer "+adminToken)
	handler.ServeHTTP(resolveAgain, request)
	if resolveAgain.Code != http.StatusConflict {
		t.Fatalf("duplicate report resolution status = %d, body = %s", resolveAgain.Code, resolveAgain.Body.String())
	}

	updateProfile := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPut, "/api/developer/me", strings.NewReader(`{"bio":"独立开发者","location":"上海","engines":["Godot 4","Unity","godot 4"]}`))
	request.Header.Set("Authorization", "Bearer "+memberToken)
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(updateProfile, request)
	if updateProfile.Code != http.StatusOK || !strings.Contains(updateProfile.Body.String(), `"bio":"独立开发者"`) || !strings.Contains(updateProfile.Body.String(), `"location":"上海"`) {
		t.Fatalf("profile update failed: %d %s", updateProfile.Code, updateProfile.Body.String())
	}
	if strings.Count(updateProfile.Body.String(), `"Godot 4"`) != 1 {
		t.Fatalf("profile engines were not normalized: %s", updateProfile.Body.String())
	}

	game, err := a.catalog.AddGame("@member", "可编辑作品", "简介", "HTML5", "休闲", "MIT")
	if err != nil {
		t.Fatal(err)
	}
	updateReviewing := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPatch, "/api/developer/games/"+catalog.IDString(game.ID), strings.NewReader(`{"title":"不应修改"}`))
	request.Header.Set("Authorization", "Bearer "+memberToken)
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(updateReviewing, request)
	if updateReviewing.Code != http.StatusConflict {
		t.Fatalf("reviewing game update status = %d, body = %s", updateReviewing.Code, updateReviewing.Body.String())
	}

	otherOwner := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPatch, "/api/developer/games/"+catalog.IDString(game.ID), strings.NewReader(`{"title":"越权修改"}`))
	request.Header.Set("Authorization", "Bearer "+adminToken)
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(otherOwner, request)
	if otherOwner.Code != http.StatusForbidden {
		t.Fatalf("non-owner game update status = %d, body = %s", otherOwner.Code, otherOwner.Body.String())
	}

	if _, err := a.catalog.ReviewGame(game.ID, "rejected", "需要补充说明"); err != nil {
		t.Fatal(err)
	}
	updateRejected := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPatch, "/api/developer/games/"+catalog.IDString(game.ID), strings.NewReader(`{"title":"已修改作品","description":"详细介绍"}`))
	request.Header.Set("Authorization", "Bearer "+memberToken)
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(updateRejected, request)
	if updateRejected.Code != http.StatusOK || !strings.Contains(updateRejected.Body.String(), `"title":"已修改作品"`) {
		t.Fatalf("rejected game update failed: %d %s", updateRejected.Code, updateRejected.Body.String())
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
	if recorder.Code != http.StatusUnprocessableEntity || !strings.Contains(recorder.Body.String(), "coverFile: file is required") {
		t.Fatalf("invalid upload status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestGameUploadRequiresCoverAndBuild(t *testing.T) {
	a := testApp()
	handler := testHandler(a)
	token := registerTestUser(t, handler)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("title", "缺少文件")
	_ = writer.WriteField("summary", "缺少必需文件")
	_ = writer.Close()

	request := httptest.NewRequest(http.MethodPost, "/api/games", &body)
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnprocessableEntity || !strings.Contains(recorder.Body.String(), `"errorCode":"VALIDATION_ERROR"`) || !strings.Contains(recorder.Body.String(), "coverFile: file is required") {
		t.Fatalf("required upload status = %d, body = %s", recorder.Code, recorder.Body.String())
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
