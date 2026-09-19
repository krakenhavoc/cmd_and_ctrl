package ws

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// TestEvictUserSessionsClosesOnlyThatUsersOlderSockets pins ADR 0051
// decision 6 at the hub: a revoked user's open sockets close with a
// terminal 1000 "session revoked"; a socket opened with a newer session,
// another user's socket and a socket with no user stay up.
func TestEvictUserSessionsClosesOnlyThatUsersOlderSockets(t *testing.T) {
	hub := NewHub(slog.New(slog.NewTextHandler(io.Discard, nil)))
	// The binding comes straight from the query string, so each dial
	// picks its own user and issue time. No game: a ping-only socket is
	// enough to watch it close.
	hub.SetAuthorizer(UpgradeAuthorizerFunc(func(r *http.Request) (Binding, error) {
		q := r.URL.Query()
		var b Binding
		if raw := q.Get("user"); raw != "" {
			b.UserID = uuid.MustParse(raw)
		}
		ms, _ := strconv.ParseInt(q.Get("iat"), 10, 64)
		b.IssuedAt = time.UnixMilli(ms).UTC()
		return b, nil
	}))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", hub.ServeWS)
	srv := httptest.NewServer(mux)
	defer srv.Close()
	base := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"

	alice, bob := uuid.New(), uuid.New()
	revokedAt := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	open := func(user uuid.UUID, iat time.Time) *websocket.Conn {
		u := base + "?iat=" + strconv.FormatInt(iat.UnixMilli(), 10)
		if user != uuid.Nil {
			u += "&user=" + user.String()
		}
		c := dial(t, u)
		t.Cleanup(func() { _ = c.Close() })
		return c
	}
	aliceOld := open(alice, revokedAt.Add(-time.Hour))
	aliceAtWatermark := open(alice, revokedAt)
	aliceNew := open(alice, revokedAt.Add(time.Millisecond))
	bobOld := open(bob, revokedAt.Add(-time.Hour))
	guest := open(uuid.Nil, revokedAt.Add(-time.Hour))

	deadline := time.Now().Add(time.Second)
	for hub.Count() < 5 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if n := hub.Count(); n != 5 {
		t.Fatalf("hub count %d, want 5", n)
	}

	if n := hub.EvictUserSessions(uuid.Nil, revokedAt); n != 0 {
		t.Errorf("EvictUserSessions(uuid.Nil) closed %d, want 0", n)
	}
	if n := hub.EvictUserSessions(alice, revokedAt); n != 2 {
		t.Errorf("EvictUserSessions closed %d, want 2", n)
	}

	for name, c := range map[string]*websocket.Conn{"old": aliceOld, "at watermark": aliceAtWatermark} {
		_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, _, err := c.ReadMessage()
		var ce *websocket.CloseError
		if !errors.As(err, &ce) || ce.Code != websocket.CloseNormalClosure || ce.Text != SessionRevokedReason {
			t.Errorf("alice's %s socket: err = %v, want close 1000 %q", name, err, SessionRevokedReason)
		}
	}

	for name, c := range map[string]*websocket.Conn{"alice new": aliceNew, "bob": bobOld, "guest": guest} {
		sendFrame(t, c, protocol.Frame{V: protocol.Version, Kind: protocol.KindPing, ID: "p", Payload: json.RawMessage(`{"msg":"still here"}`)})
		if f := readFrame(t, c); f.Kind != protocol.KindPong {
			t.Errorf("%s socket: got %q, want pong", name, f.Kind)
		}
	}
}
