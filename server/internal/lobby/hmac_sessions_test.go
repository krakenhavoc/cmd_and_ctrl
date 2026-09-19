package lobby

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
)

// hmac_sessions_test.go runs the lobby's session paths against the
// stateless authenticator production uses (#517, ADR 0044 decision
// 3). The rest of the lobby suite runs on auth.MemoryAuthenticator;
// these are the places where the two could differ.

const lobbyTestSessionKey = "lobby-test-session-key-0123456789abcdef"

func newHMAC(t *testing.T) *auth.HMACAuthenticator {
	t.Helper()
	a, err := auth.NewHMACAuthenticator([]byte(lobbyTestSessionKey))
	if err != nil {
		t.Fatalf("NewHMACAuthenticator: %v", err)
	}
	return a
}

// TestPlayerSessionSurvivesARestartedAuthenticator: a seat claimed on
// one process opens a socket on the same process, and is accepted by
// the WS authorizer of a process started later with the same key —
// the deploy, minus the room restore that persist_test.go covers.
func TestPlayerSessionSurvivesARestartedAuthenticator(t *testing.T) {
	srv, l := newTestHTTPStackWithAuth(t, nil, newHMAC(t))
	meta, err := l.Create("FNM")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Alice"})
	var joined sessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&joined); err != nil {
		t.Fatalf("decode join: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK || joined.Token == "" {
		t.Fatalf("join: status %d, token %q", resp.StatusCode, joined.Token)
	}

	// The signed token goes into the WS URL verbatim.
	if url.QueryEscape(joined.Token) != joined.Token {
		t.Errorf("session token needs escaping in ?token=: %q", joined.Token)
	}
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws?token=" + joined.Token + "&game=" + meta.ID.String()
	conn, resp2, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		body := ""
		if resp2 != nil {
			b, _ := io.ReadAll(resp2.Body)
			body = string(b)
		}
		t.Fatalf("ws dial: %v (body=%s)", err, body)
	}
	conn.Close()

	// "After the deploy": nothing shared with the first authenticator
	// but the key.
	restarted := &WSAuthorizer{Auth: newHMAC(t)}
	req := httptest.NewRequest(http.MethodGet, "/ws?token="+joined.Token+"&game="+meta.ID.String(), nil)
	bind, err := restarted.AuthorizeUpgrade(req)
	if err != nil {
		t.Fatalf("restarted authorizer rejected a pre-restart session: %v", err)
	}
	if bind.GameID != meta.ID || bind.PlayerID != joined.PlayerID || bind.PlayerID == uuid.Nil {
		t.Errorf("binding: got %+v, want game %s player %s", bind, meta.ID, joined.PlayerID)
	}
}

// TestIdentitySessionTradesInUnderHMAC walks ADR 0050's login-first
// path with signed tokens: an identity session minted by one process
// is traded at POST /join on a process that never saw it issued, and
// the Discord fields reach the seat.
func TestIdentitySessionTradesInUnderHMAC(t *testing.T) {
	issuer := newHMAC(t)
	idTok, _, err := issuer.Issue(context.Background(), auth.Principal{
		Role:              auth.RoleIdentified,
		Name:              "Alice",
		DiscordID:         "discord-99",
		DiscordUsername:   "alice",
		DiscordGlobalName: "Alice",
		DiscordAvatarHash: "avhash",
	}, time.Hour)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	srv, l := newTestHTTPStackWithAuth(t, nil, newHMAC(t))
	meta, err := l.Create("FNM")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	resp := postJSON(t, srv, "/join", idTok, map[string]string{"invite_token": meta.InviteToken})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d, want 200", resp.StatusCode)
	}
	var joined sessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&joined); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if joined.Principal.Role != auth.RolePlayer || joined.Principal.DiscordID != "discord-99" {
		t.Errorf("principal: got %+v, want a player carrying discord-99", joined.Principal)
	}
	updated, err := l.Get(meta.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(updated.Players) != 1 || updated.Players[0].DiscordAvatarHash != "avhash" {
		t.Errorf("seat did not carry the Discord identity: %+v", updated.Players)
	}
}

// TestLogoutWithStatelessSessions pins what POST /logout means under
// HMAC sessions. It still answers 204 and clears the cookie, which is
// what ends the session in the browser. It cannot kill the token
// itself: Revoke is advisory (ADR 0044 decision 3), so a copy of the
// token held elsewhere keeps validating until it expires. Compare
// TestLogoutRevokesSession, which runs on MemoryAuthenticator.
func TestLogoutWithStatelessSessions(t *testing.T) {
	a := newHMAC(t)
	srv, _ := newTestHTTPStackWithAuth(t, nil, a)

	resp := postJSON(t, srv, "/admin/login", "", adminLoginRequest{Token: "shared-admin-token"})
	var s sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&s)
	resp.Body.Close()
	if s.Token == "" {
		t.Fatalf("admin login returned no token (status %d)", resp.StatusCode)
	}

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/logout", nil)
	req.Header.Set("Authorization", "Bearer "+s.Token)
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("logout: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("logout status: got %d, want 204", resp.StatusCode)
	}
	var sawClear bool
	for _, c := range resp.Cookies() {
		if c.Name == auth.SessionCookie && c.MaxAge < 0 {
			sawClear = true
		}
	}
	if !sawClear {
		t.Error("logout did not emit a cookie-clearing Set-Cookie")
	}

	if _, err := a.Validate(context.Background(), s.Token); err != nil {
		t.Errorf("advisory revoke: the token should still validate until expiry, got %v", err)
	}
}
