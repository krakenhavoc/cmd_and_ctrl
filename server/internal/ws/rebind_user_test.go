package ws

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// TestRebindUserSessionsClosesOnlyTheWrongAdminBit pins ADR 0112 §2
// item 5 at the hub: a switch of admin mode closes, with the
// non-terminal 4001, exactly the user's sockets whose admin bit is now
// wrong. Their other sockets, another user's and a socket with no user
// stay up. The bit is never flipped in place.
func TestRebindUserSessionsClosesOnlyTheWrongAdminBit(t *testing.T) {
	hub := NewHub(slog.New(slog.NewTextHandler(io.Discard, nil)))
	hub.SetAuthorizer(UpgradeAuthorizerFunc(func(r *http.Request) (Binding, error) {
		q := r.URL.Query()
		var b Binding
		if raw := q.Get("user"); raw != "" {
			b.UserID = uuid.MustParse(raw)
		}
		b.Admin = q.Get("admin") == "1"
		return b, nil
	}))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", hub.ServeWS)
	srv := httptest.NewServer(mux)
	defer srv.Close()
	base := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"

	alice, bob := uuid.New(), uuid.New()
	open := func(user uuid.UUID, admin bool) *websocket.Conn {
		u := base + "?x=1"
		if user != uuid.Nil {
			u += "&user=" + user.String()
		}
		if admin {
			u += "&admin=1"
		}
		c := dial(t, u)
		t.Cleanup(func() { _ = c.Close() })
		return c
	}
	aliceAdmin := open(alice, true)
	alicePlayer := open(alice, false)
	bobAdmin := open(bob, true)
	token := open(uuid.Nil, true)

	deadline := time.Now().Add(time.Second)
	for hub.Count() < 4 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if n := hub.Count(); n != 4 {
		t.Fatalf("hub count %d, want 4", n)
	}

	if n := hub.RebindUserSessions(uuid.Nil, false); n != 0 {
		t.Errorf("RebindUserSessions(uuid.Nil) closed %d, want 0 (the token has no mode)", n)
	}
	// Alice switches to player mode: only her admin socket is wrong.
	if n := hub.RebindUserSessions(alice, false); n != 1 {
		t.Errorf("RebindUserSessions(alice, false) closed %d, want 1", n)
	}
	assertClosed4001(t, "alice's admin socket", aliceAdmin)
	for name, c := range map[string]*websocket.Conn{"alice's player socket": alicePlayer, "bob": bobAdmin, "the token": token} {
		sendFrame(t, c, protocol.Frame{V: protocol.Version, Kind: protocol.KindPing, ID: "p", Payload: json.RawMessage(`{"msg":"still here"}`)})
		if f := readFrame(t, c); f.Kind != protocol.KindPong {
			t.Errorf("%s: got %q, want pong", name, f.Kind)
		}
	}

	// And back to admin mode: now her player socket is the wrong one.
	if n := hub.RebindUserSessions(alice, true); n != 1 {
		t.Errorf("RebindUserSessions(alice, true) closed %d, want 1", n)
	}
	assertClosed4001(t, "alice's player socket", alicePlayer)
}

func assertClosed4001(t *testing.T, name string, c *websocket.Conn) {
	t.Helper()
	_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
	for {
		_, _, err := c.ReadMessage()
		if err == nil {
			continue // a snapshot or pong already in flight
		}
		var ce *websocket.CloseError
		if !errors.As(err, &ce) || ce.Code != AdminModeChangedCode || ce.Text != AdminModeChangedReason {
			t.Errorf("%s: err = %v, want close %d %q", name, err, AdminModeChangedCode, AdminModeChangedReason)
		}
		return
	}
}
