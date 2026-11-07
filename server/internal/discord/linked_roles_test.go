package discord

// Tests for linked_roles.go (ADR 0163): the configuration, the metadata
// registration, the per-person PUT and the role-connection authorize
// URL, against an httptest.Server standing in for Discord.

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

const testAppID = "777777777777777777"

func TestLinkedRolesFromEnv(t *testing.T) {
	cases := []struct {
		name        string
		vals        map[string]string
		wantErr     string
		wantEnabled bool
		wantMissing string
	}{
		{name: "unset is off", vals: map[string]string{AppIDEnv: testAppID, BotTokenEnv: "tok"}},
		{name: "0 is off", vals: map[string]string{LinkedRolesEnv: "0", AppIDEnv: testAppID, BotTokenEnv: "tok"}},
		{name: "1 with the id and the token is on",
			vals: map[string]string{LinkedRolesEnv: " 1 ", AppIDEnv: " " + testAppID + " ", BotTokenEnv: "tok"}, wantEnabled: true},
		{name: "1 without the token is off but boots",
			vals: map[string]string{LinkedRolesEnv: "1", AppIDEnv: testAppID}, wantMissing: BotTokenEnv},
		{name: "1 without the id is off but boots",
			vals: map[string]string{LinkedRolesEnv: "1", BotTokenEnv: "tok"}, wantMissing: AppIDEnv},
		{name: "another switch value fails", vals: map[string]string{LinkedRolesEnv: "true"}, wantErr: LinkedRolesEnv},
		{name: "a malformed id fails", vals: map[string]string{LinkedRolesEnv: "1", AppIDEnv: "my-app"}, wantErr: AppIDEnv},
	}
	for _, tc := range cases {
		lr, err := LinkedRolesFromEnv(envOf(tc.vals))
		if tc.wantErr != "" {
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("%s: err %v, want one naming %s", tc.name, err, tc.wantErr)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: %v", tc.name, err)
			continue
		}
		if lr.Enabled() != tc.wantEnabled {
			t.Errorf("%s: Enabled() = %v, want %v (%+v)", tc.name, lr.Enabled(), tc.wantEnabled, lr)
		}
		if !strings.Contains(lr.Missing(), tc.wantMissing) || (tc.wantMissing == "" && lr.Missing() != "") {
			t.Errorf("%s: Missing() = %q, want %q", tc.name, lr.Missing(), tc.wantMissing)
		}
	}
}

// discordCall is one request the stub saw.
type discordCall struct {
	Method, Path, Auth string
	Body               []byte
}

// linkedRolesStub is Discord's REST API: it records each call and
// answers with the next status in replies (200 when they run out).
type linkedRolesStub struct {
	mu      sync.Mutex
	calls   []discordCall
	replies []int
}

func newLinkedRolesStub(t *testing.T, replies ...int) *linkedRolesStub {
	t.Helper()
	s := &linkedRolesStub{replies: replies}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		s.mu.Lock()
		s.calls = append(s.calls, discordCall{Method: r.Method, Path: r.URL.Path, Auth: r.Header.Get("Authorization"), Body: body})
		status := http.StatusOK
		if len(s.replies) > 0 {
			status, s.replies = s.replies[0], s.replies[1:]
		}
		s.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if status == http.StatusTooManyRequests {
			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"message":"You are being rate limited.","retry_after":1.5,"global":false}`))
			return
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)
	orig := botAPIBase
	botAPIBase = srv.URL
	t.Cleanup(func() { botAPIBase = orig })
	return s
}

func (s *linkedRolesStub) all() []discordCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]discordCall(nil), s.calls...)
}

func onLR() LinkedRoles { return LinkedRoles{AppID: testAppID, Token: "bot-secret", Requested: true} }

func TestRegisterMetadataSendsTheThreeFieldsWithTheBotToken(t *testing.T) {
	stub := newLinkedRolesStub(t)
	if err := onLR().RegisterMetadata(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	calls := stub.all()
	if len(calls) != 1 {
		t.Fatalf("%d calls, want 1", len(calls))
	}
	c := calls[0]
	if c.Method != http.MethodPut || c.Path != "/applications/"+testAppID+"/role-connections/metadata" || c.Auth != "Bot bot-secret" {
		t.Errorf("call = %s %s auth %q", c.Method, c.Path, c.Auth)
	}
	var fields []RoleConnectionMetadata
	if err := json.Unmarshal(c.Body, &fields); err != nil {
		t.Fatal(err)
	}
	want := map[string]int{MetaGamesPlayed: 2, MetaGamesWon: 2, MetaJoined: 5}
	if len(fields) != len(want) {
		t.Fatalf("fields = %+v, want exactly %v", fields, want)
	}
	for _, f := range fields {
		if want[f.Key] != f.Type || f.Name == "" || f.Description == "" || len(f.Description) > 200 {
			t.Errorf("field %+v", f)
		}
	}
}

func TestRegisterMetadataOffTouchesNothing(t *testing.T) {
	stub := newLinkedRolesStub(t)
	if err := (LinkedRoles{AppID: testAppID}).RegisterMetadata(context.Background(), nil); !errors.Is(err, ErrLinkedRolesOff) {
		t.Errorf("off: %v, want ErrLinkedRolesOff", err)
	}
	if n := len(stub.all()); n != 0 {
		t.Errorf("off made %d calls", n)
	}
}

func TestPutRoleConnectionWritesTheNumbersWithThePersonsToken(t *testing.T) {
	stub := newLinkedRolesStub(t)
	rc := RoleConnection{
		PlatformName: RoleConnectionPlatform, PlatformUsername: "Alice",
		GamesPlayed: 12, GamesWon: 3, Joined: time.Date(2026, 9, 1, 12, 0, 0, 0, time.FixedZone("x", 3600)),
	}
	if err := onLR().PutRoleConnection(context.Background(), nil, "person-access", rc); err != nil {
		t.Fatal(err)
	}
	c := stub.all()[0]
	if c.Method != http.MethodPut || c.Path != "/users/@me/applications/"+testAppID+"/role-connection" || c.Auth != "Bearer person-access" {
		t.Errorf("call = %s %s auth %q", c.Method, c.Path, c.Auth)
	}
	var body struct {
		PlatformName     string            `json:"platform_name"`
		PlatformUsername string            `json:"platform_username"`
		Metadata         map[string]string `json:"metadata"`
	}
	if err := json.Unmarshal(c.Body, &body); err != nil {
		t.Fatal(err)
	}
	if body.PlatformName != "CMD & CTRL" || body.PlatformUsername != "Alice" {
		t.Errorf("body = %+v", body)
	}
	want := map[string]string{MetaGamesPlayed: "12", MetaGamesWon: "3", MetaJoined: "2026-09-01T11:00:00Z"}
	if len(body.Metadata) != len(want) {
		t.Errorf("metadata = %v, want %v", body.Metadata, want)
	}
	for k, v := range want {
		if body.Metadata[k] != v {
			t.Errorf("metadata[%s] = %q, want %q", k, body.Metadata[k], v)
		}
	}
	// Nothing else about the person rides along.
	if strings.Contains(string(c.Body), "bot-secret") {
		t.Error("the bot token is in the body")
	}
}

func TestPutRoleConnectionTruncatesALongName(t *testing.T) {
	stub := newLinkedRolesStub(t)
	long := strings.Repeat("é", MaxPlatformUsername+20)
	if err := onLR().PutRoleConnection(context.Background(), nil, "a", RoleConnection{PlatformUsername: long}); err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	_ = json.Unmarshal(stub.all()[0].Body, &body)
	if got := []rune(body["platform_username"].(string)); len(got) != MaxPlatformUsername {
		t.Errorf("platform_username is %d runes, want %d", len(got), MaxPlatformUsername)
	}
}

func TestClearRoleConnectionSendsNothingAboutThePerson(t *testing.T) {
	stub := newLinkedRolesStub(t)
	if err := onLR().ClearRoleConnection(context.Background(), nil, "a"); err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err := json.Unmarshal(stub.all()[0].Body, &body); err != nil {
		t.Fatal(err)
	}
	if body["platform_name"] != nil || body["platform_username"] != nil {
		t.Errorf("clear body = %v, want null names", body)
	}
	if m, ok := body["metadata"].(map[string]any); !ok || len(m) != 0 {
		t.Errorf("clear metadata = %v, want {}", body["metadata"])
	}
}

func TestPutRoleConnectionErrors(t *testing.T) {
	for _, tc := range []struct {
		status int
		check  func(error) bool
		want   string
	}{
		{http.StatusTooManyRequests, func(err error) bool {
			var rl *RateLimitError
			return errors.As(err, &rl) && rl.RetryAfter == 1500*time.Millisecond && Transient(err)
		}, "a *RateLimitError with retry_after 1.5s"},
		{http.StatusBadGateway, func(err error) bool { return errors.Is(err, ErrVoiceUnavailable) && Transient(err) }, "unavailable"},
		{http.StatusUnauthorized, func(err error) bool { return errors.Is(err, ErrRoleConnectionUnauthorized) && !Transient(err) }, "unauthorized"},
		{http.StatusBadRequest, func(err error) bool { return err != nil && !Transient(err) }, "a plain error"},
	} {
		newLinkedRolesStub(t, tc.status)
		err := onLR().PutRoleConnection(context.Background(), nil, "secret-access", RoleConnection{})
		if !tc.check(err) {
			t.Errorf("status %d: %v, want %s", tc.status, err, tc.want)
		}
		if err != nil && strings.Contains(err.Error(), "secret-access") {
			t.Errorf("status %d: the access token is in the error", tc.status)
		}
	}
}

func TestRoleConnectionAuthorizeURL(t *testing.T) {
	c := Config{ClientID: "cid", ClientSecret: "s", RedirectURI: "https://cmd.example/auth/discord/callback",
		Gate: Gate{GuildID: "1", RoleID: "2"}}
	u, err := url.Parse(c.RoleConnectionAuthorizeURL("st", "ch"))
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if q.Get("scope") != "identify role_connections.write" {
		t.Errorf("scope = %q; the role-connection grant asks for identify and role_connections.write only, gate or not", q.Get("scope"))
	}
	if q.Get("redirect_uri") != c.RedirectURI || q.Get("prompt") != "consent" || q.Get("state") != "st" ||
		q.Get("code_challenge") != "ch" || q.Get("client_id") != "cid" {
		t.Errorf("query = %v", q)
	}
	// Sign-in's own URL is unchanged by the refactor.
	if s, _ := url.Parse(c.AuthorizeURL("st", "ch", PromptNone)); s.Query().Get("scope") != gatedScopes {
		t.Errorf("sign-in scope = %q, want %q", s.Query().Get("scope"), gatedScopes)
	}
}

func TestTokenResponseHasScope(t *testing.T) {
	tr := TokenResponse{Scope: "identify role_connections.write"}
	if !tr.HasScope(RoleConnectionScope) || !tr.HasScope("identify") || tr.HasScope("role_connections") {
		t.Errorf("HasScope wrong for %q", tr.Scope)
	}
	if (TokenResponse{}).HasScope(RoleConnectionScope) {
		t.Error("an empty scope has role_connections.write")
	}
}

func TestStartRoleConnectionIsItsOwnFlow(t *testing.T) {
	s := NewStateStore()
	st, ch, err := s.StartRoleConnection()
	if err != nil || st == "" || ch == "" {
		t.Fatalf("start: %q %q %v", st, ch, err)
	}
	e, err := s.Consume(st)
	if err != nil {
		t.Fatal(err)
	}
	if !e.RoleConnection || e.Unbound() || e.Link() || e.GameID != uuid.Nil || e.Prompt != PromptConsent {
		t.Errorf("entry = %+v: want a consent role-connection round that is neither unbound nor a link", e)
	}
	// Consent rounds are never retried, so a refusal is shown.
	if _, _, err := s.Retry(e); !errors.Is(err, ErrNoRetry) {
		t.Errorf("Retry = %v, want ErrNoRetry", err)
	}
}
