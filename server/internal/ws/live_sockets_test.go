package ws

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
)

// Hub.LiveSockets, the admin views' copy of the live connections (ADR
// 0124 §5): what it copies, that it agrees with MetricsSockets, and
// that it takes the hub's read lock only.

// liveSocketsHub serves sockets whose binding is read from the query
// string (?player=, ?user=, ?ro=1, ?admin=1), with no game, on a hub
// whose clock is fixed at `at`.
func liveSocketsHub(t *testing.T, at time.Time) (*Hub, string) {
	t.Helper()
	hub := NewHub(slog.New(slog.NewTextHandler(io.Discard, nil)))
	hub.clock = func() time.Time { return at }
	hub.SetAuthorizer(UpgradeAuthorizerFunc(func(r *http.Request) (Binding, error) {
		q := r.URL.Query()
		var b Binding
		if raw := q.Get("player"); raw != "" {
			b.PlayerID = uuid.MustParse(raw)
		}
		if raw := q.Get("user"); raw != "" {
			b.UserID = uuid.MustParse(raw)
		}
		b.ReadOnly = q.Get("ro") == "1"
		b.Admin = q.Get("admin") == "1"
		return b, nil
	}))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", hub.ServeWS)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return hub, "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
}

func waitForHubCount(t *testing.T, hub *Hub, n int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for hub.Count() != n && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if got := hub.Count(); got != n {
		t.Fatalf("hub count %d, want %d", got, n)
	}
}

func TestLiveSocketsCopiesEveryConnection(t *testing.T) {
	at := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	hub, base := liveSocketsHub(t, at)
	seat, user := uuid.New(), uuid.New()
	for _, q := range []string{
		"?player=" + seat.String() + "&user=" + user.String(),
		"?ro=1",
		"?admin=1&user=" + user.String(),
	} {
		c := dial(t, base+q)
		t.Cleanup(func() { _ = c.Close() })
	}
	waitForHubCount(t, hub, 3)

	got := hub.LiveSockets()
	byRole := map[string]LiveSocket{}
	for _, s := range got {
		if s.ConnectedAt != at {
			t.Errorf("%s socket ConnectedAt = %v, want the hub's clock %v", s.Role(), s.ConnectedAt, at)
		}
		byRole[s.Role()] = s
	}
	if s := byRole[metrics.RoleSeat]; s.PlayerID != seat || s.UserID != user || s.ReadOnly || s.Admin {
		t.Errorf("seat socket = %+v", s)
	}
	if s := byRole[metrics.RoleSpectator]; s.PlayerID != uuid.Nil || s.UserID != uuid.Nil || !s.ReadOnly || s.Admin {
		t.Errorf("spectator socket = %+v", s)
	}
	if s := byRole[metrics.RoleAdmin]; s.PlayerID != uuid.Nil || s.UserID != user || s.ReadOnly || !s.Admin {
		t.Errorf("admin socket = %+v", s)
	}

	// Stripped to the metrics type, it is exactly what MetricsSockets
	// copies.
	key := func(s metrics.Socket) string { return s.Role + string(s.Game[:]) + string(s.Player[:]) }
	var fromLive, fromMetrics []string
	for _, s := range got {
		fromLive = append(fromLive, key(s.Metrics()))
	}
	for _, s := range hub.MetricsSockets() {
		fromMetrics = append(fromMetrics, key(s))
	}
	sort.Strings(fromLive)
	sort.Strings(fromMetrics)
	if strings.Join(fromLive, "|") != strings.Join(fromMetrics, "|") {
		t.Errorf("LiveSockets stripped = %q, MetricsSockets = %q", fromLive, fromMetrics)
	}
}

// LiveSockets takes the hub's read lock and nothing stronger: it
// answers while another reader holds the hub, and the hub is free when
// it returns.
func TestLiveSocketsTakesOnlyTheHubsReadLock(t *testing.T) {
	hub, base := liveSocketsHub(t, time.Now())
	c := dial(t, base+"?ro=1")
	t.Cleanup(func() { _ = c.Close() })
	waitForHubCount(t, hub, 1)

	hub.mu.RLock()
	done := make(chan []LiveSocket, 1)
	go func() { done <- hub.LiveSockets() }()
	select {
	case got := <-done:
		if len(got) != 1 {
			t.Errorf("LiveSockets beside a reader = %d sockets, want 1", len(got))
		}
	case <-time.After(5 * time.Second):
		t.Error("LiveSockets blocked while another reader held the hub: it takes the write lock")
	}
	hub.mu.RUnlock()

	if !hub.mu.TryLock() {
		t.Fatal("LiveSockets left the hub's lock held")
	}
	hub.mu.Unlock()
}
