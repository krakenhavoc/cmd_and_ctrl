package lobby

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// The games, players and WebSocket metrics against a real lobby, hub
// and sockets (ADR 0123 §3, §11).

// metricsStack is newTestHTTPStackWithAuth with the hub handed back,
// and a registry carrying the tables collector over the live lobby and
// hub as main.go registers it.
func metricsStack(t *testing.T) (*httptest.Server, *Lobby, *ws.Hub, *prometheus.Registry) {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	a := auth.NewMemoryAuthenticator()
	mgr := ws.NewRoomManager(log, "")
	l := NewLobby(mgr)
	hub := ws.NewHub(log)
	hub.SetManager(mgr)
	hub.SetAuthorizer(&WSAuthorizer{Auth: a})
	l.SetStateBroadcaster(hub)

	mux := http.NewServeMux()
	mux.Handle("/", Handler(Config{Lobby: l, Auth: a, AdminToken: "shared-admin-token"}))
	mux.HandleFunc("GET /ws", hub.ServeWS)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	reg := prometheus.NewRegistry()
	reg.MustRegister(metrics.NewTablesCollector(l, hub))
	return srv, l, hub, reg
}

// series gathers g into "name{l=v,…}" → value, labels sorted; a
// histogram reports its sample count.
func series(t *testing.T, g prometheus.Gatherer) map[string]float64 {
	t.Helper()
	families, err := g.Gather()
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]float64{}
	for _, f := range families {
		for _, m := range f.GetMetric() {
			var ls []string
			for _, lp := range m.GetLabel() {
				ls = append(ls, lp.GetName()+"="+lp.GetValue())
			}
			sort.Strings(ls)
			var v float64
			switch {
			case m.GetCounter() != nil:
				v = m.GetCounter().GetValue()
			case m.GetGauge() != nil:
				v = m.GetGauge().GetValue()
			case m.GetHistogram() != nil:
				v = float64(m.GetHistogram().GetSampleCount())
			}
			out[f.GetName()+"{"+strings.Join(ls, ",")+"}"] = v
		}
	}
	return out
}

// event reads one series of metrics.Registry: the package-level event
// counters.
func event(t *testing.T, name string) float64 {
	t.Helper()
	v, ok := series(t, metrics.Registry)[name]
	if !ok {
		t.Fatalf("metrics.Registry has no series %s", name)
	}
	return v
}

// session posts body to path and decodes the session it answers.
// It returns an error rather than failing the test, so the traffic
// goroutines of TestScrapeDuringTableTraffic can call it.
func session(srv *httptest.Server, path string, body any) (sessionResponse, error) {
	raw, err := json.Marshal(body)
	if err != nil {
		return sessionResponse{}, err
	}
	resp, err := srv.Client().Post(srv.URL+path, "application/json", strings.NewReader(string(raw)))
	if err != nil {
		return sessionResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return sessionResponse{}, fmt.Errorf("POST %s: %d", path, resp.StatusCode)
	}
	var s sessionResponse
	err = json.NewDecoder(resp.Body).Decode(&s)
	return s, err
}

// joinOverHTTP seats name through POST /games/{id}/join and returns
// the seat's session token and player ID.
func joinOverHTTP(srv *httptest.Server, meta GameMeta, name string) (string, uuid.UUID, error) {
	s, err := session(srv, "/games/"+meta.ID.String()+"/join", joinRequest{InviteToken: meta.InviteToken, Name: name})
	return s.Token, s.PlayerID, err
}

// spectateOverHTTP opens a spectator session on the table.
func spectateOverHTTP(srv *httptest.Server, meta GameMeta) (string, error) {
	s, err := session(srv, "/games/"+meta.ID.String()+"/spectate",
		spectateRequest{InviteToken: meta.SpectatorInvite, Name: "watcher"})
	return s.Token, err
}

// dialAdmitted opens a socket with token and reads its first frame,
// the initial snapshot. The hub stages it before it admits the
// connection, and starts the pump that writes it only after, so once
// it is read the hub holds and has counted the socket.
func dialAdmitted(srv *httptest.Server, token string) (*websocket.Conn, error) {
	conn, resp, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http")+"/ws?token="+token, nil)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("ws dial: %w", err)
	}
	if _, _, err := conn.ReadMessage(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("read initial snapshot: %w", err)
	}
	return conn, nil
}

// mustJoin is joinOverHTTP for the test's own goroutine.
func mustJoin(t *testing.T, srv *httptest.Server, meta GameMeta, name string) (string, uuid.UUID) {
	t.Helper()
	token, id, err := joinOverHTTP(srv, meta, name)
	if err != nil {
		t.Fatalf("join %s: %v", name, err)
	}
	return token, id
}

// mustDial is dialAdmitted for the test's own goroutine, closed at the
// end of the test.
func mustDial(t *testing.T, srv *httptest.Server, token string) *websocket.Conn {
	t.Helper()
	conn, err := dialAdmitted(srv, token)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

// staticUsers is a database with no users, for the users collector in
// the traffic test, which is about the lobby half.
type staticUsers struct{}

func (staticUsers) CountUsers(context.Context) (int, error) { return 0, nil }

func (staticUsers) LastPlayed(context.Context, time.Time) (map[string]time.Time, error) {
	return map[string]time.Time{}, nil
}

// ADR 0123 §11: a built table — two human seats, one of them a guest
// with a live socket, a bot seat, and a spectator — gives the expected
// cmdctrl_seats and cmdctrl_seats_connected series.
func TestMetricsOfABuiltTable(t *testing.T) {
	srv, l, hub, reg := metricsStack(t)

	meta, err := l.Create("Metrics")
	if err != nil {
		t.Fatal(err)
	}
	aliceToken, alice := mustJoin(t, srv, meta, "Alice")
	// Bob is signed in and seated, with no socket.
	_, bob, err := l.JoinAs(meta.ID, meta.InviteToken, "Bob", DiscordIdentity{}, uuid.New())
	if err != nil {
		t.Fatalf("JoinAs Bob: %v", err)
	}
	if _, _, err := l.AddBot(meta.ID, "Bot 1", "random", "", "Mono Red", botDeck(20)); err != nil {
		t.Fatalf("AddBot: %v", err)
	}
	for _, p := range []uuid.UUID{alice, bob} {
		if _, err := l.SetDeck(meta.ID, p, "deck", botDeck(20)); err != nil {
			t.Fatalf("SetDeck: %v", err)
		}
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	mustDial(t, srv, aliceToken)
	spec, err := spectateOverHTTP(srv, meta)
	if err != nil {
		t.Fatal(err)
	}
	mustDial(t, srv, spec)
	if hub.Count() != 2 {
		t.Fatalf("hub has %d sockets, want 2", hub.Count())
	}

	got := series(t, reg)
	for k, want := range map[string]float64{
		"cmdctrl_games{archived=false,state=active}": 1,
		"cmdctrl_games{archived=false,state=lobby}":  0,
		"cmdctrl_practice_games{}":                   0,

		"cmdctrl_seats{kind=human}": 2,
		"cmdctrl_seats{kind=bot}":   1,
		"cmdctrl_seats{kind=agent}": 0,

		"cmdctrl_seats_connected{account=guest,kind=human}":     1,
		"cmdctrl_seats_connected{account=signed_in,kind=human}": 0,
		"cmdctrl_seats_connected{account=guest,kind=agent}":     0,
		"cmdctrl_seats_connected{account=signed_in,kind=agent}": 0,

		"cmdctrl_spectators_connected{}":         1,
		"cmdctrl_ws_connections{role=seat}":      1,
		"cmdctrl_ws_connections{role=spectator}": 1,
		"cmdctrl_ws_connections{role=admin}":     0,
	} {
		if got[k] != want {
			t.Errorf("%s = %v, want %v", k, got[k], want)
		}
		if _, ok := got[k]; !ok {
			t.Errorf("%s: no such series", k)
		}
	}

	// Archiving the table closes it: its seats are no longer at an
	// active table.
	if _, err := l.SetArchived(meta.ID, true); err != nil {
		t.Fatal(err)
	}
	got = series(t, reg)
	if got["cmdctrl_games{archived=true,state=active}"] != 1 || got["cmdctrl_seats{kind=human}"] != 0 ||
		got["cmdctrl_seats_connected{account=guest,kind=human}"] != 0 {
		t.Errorf("after archive: games=%v seats=%v connected=%v", got["cmdctrl_games{archived=true,state=active}"],
			got["cmdctrl_seats{kind=human}"], got["cmdctrl_seats_connected{account=guest,kind=human}"])
	}

	if err := metrics.CheckClosedLabels(reg); err != nil {
		t.Fatalf("label rule broken:\n%v", err)
	}
}

// The live accounts are the signed-in seats at running, unarchived,
// non-practice tables.
func TestMetricsLiveUsers(t *testing.T) {
	l, _, _ := newPracticeLobby(t)
	running, _, _ := startTwoSeatGame(t, l, "Running")
	signedIn := uuid.New()
	l.mu.Lock()
	l.games[running.ID].meta.Players[0].UserID = signedIn.String()
	l.mu.Unlock()

	waiting, _ := l.Create("Waiting")
	if _, _, err := l.JoinAs(waiting.ID, waiting.InviteToken, "Carol", DiscordIdentity{}, uuid.New()); err != nil {
		t.Fatal(err)
	}
	human, bot := practiceSeats("user:p")
	human.UserID = uuid.New()
	if _, _, err := l.CreatePractice(human, bot); err != nil {
		t.Fatal(err)
	}

	if got := l.MetricsLiveUsers(); len(got) != 1 || got[0] != signedIn.String() {
		t.Errorf("MetricsLiveUsers = %v, want just %s", got, signedIn)
	}
	if _, err := l.SetArchived(running.ID, true); err != nil {
		t.Fatal(err)
	}
	if got := l.MetricsLiveUsers(); len(got) != 0 {
		t.Errorf("MetricsLiveUsers after archive = %v, want none", got)
	}
}

// Each game counter moves by exactly one per event, and each table's
// end is counted once whichever path ends it.
func TestGameCountersMoveByOne(t *testing.T) {
	srv, l, _, _ := metricsStack(t)

	const (
		created = "cmdctrl_games_created_total{}"
		started = "cmdctrl_games_started_total{seats=2}"
		win     = "cmdctrl_games_ended_total{outcome=win}"
		draw    = "cmdctrl_games_ended_total{outcome=draw}"
		closed  = "cmdctrl_games_ended_total{outcome=closed}"
		length  = "cmdctrl_game_duration_seconds{}"
	)
	names := []string{created, started, win, draw, closed, length}
	base := map[string]float64{}
	for _, n := range names {
		base[n] = event(t, n)
	}
	check := func(step string, want map[string]float64) {
		t.Helper()
		for _, n := range names {
			if got := event(t, n) - base[n]; got != want[n] {
				t.Errorf("%s: %s moved by %v, want %v", step, n, got, want[n])
			}
		}
	}

	// POST /games counts; a create that is not one does not.
	resp := postJSON(t, srv, "/admin/login", "", adminLoginRequest{Token: "shared-admin-token"})
	var admin sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&admin)
	resp.Body.Close()
	resp = postJSON(t, srv, "/games", admin.Token, createGameRequest{Name: "Over HTTP"})
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /games: %d", resp.StatusCode)
	}
	check("POST /games", map[string]float64{created: 1})

	// A start, then a concession that ends the game with a winner.
	won, alice, _ := startTwoSeatGame(t, l, "Won")
	check("start", map[string]float64{created: 1, started: 1})
	room := l.RoomOf(won.ID)
	if _, _, err := room.Apply(alice, func() error { return room.Game.Concede(alice) }); err != nil {
		t.Fatalf("concede: %v", err)
	}
	waitFor(t, func() bool { return event(t, win)-base[win] >= 1 })
	check("concede", map[string]float64{created: 1, started: 1, win: 1, length: 1})
	// Archiving and deleting an ended game ends nothing.
	if _, err := l.SetArchived(won.ID, true); err != nil {
		t.Fatal(err)
	}
	if err := l.Delete(won.ID); err != nil {
		t.Fatal(err)
	}
	check("archive and delete an ended game", map[string]float64{created: 1, started: 1, win: 1, length: 1})

	// Archiving a running table closes it, once: unarchiving and
	// archiving again, and deleting it, count nothing more.
	archived, _, _ := startTwoSeatGame(t, l, "Archived")
	if _, err := l.SetArchived(archived.ID, true); err != nil {
		t.Fatal(err)
	}
	check("archive a running table", map[string]float64{created: 1, started: 2, win: 1, closed: 1, length: 2})
	for _, flip := range []bool{false, true} {
		if _, err := l.SetArchived(archived.ID, flip); err != nil {
			t.Fatal(err)
		}
	}
	if err := l.Delete(archived.ID); err != nil {
		t.Fatal(err)
	}
	check("unarchive, archive, delete", map[string]float64{created: 1, started: 2, win: 1, closed: 1, length: 2})

	// Deleting a running table closes it.
	deleted, _, _ := startTwoSeatGame(t, l, "Deleted")
	if err := l.Delete(deleted.ID); err != nil {
		t.Fatal(err)
	}
	check("delete a running table", map[string]float64{created: 1, started: 3, win: 1, closed: 2, length: 3})

	// A table that never started neither starts nor ends.
	never, _ := l.Create("Never")
	if err := l.Delete(never.ID); err != nil {
		t.Fatal(err)
	}
	check("delete an unstarted table", map[string]float64{created: 1, started: 3, win: 1, closed: 2, length: 3})
}

// A practice table is never a started or ended game.
func TestPracticeTablesAreNotCounted(t *testing.T) {
	l, _, _ := newPracticeLobby(t)
	before := series(t, metrics.Registry)
	human, bot := practiceSeats("user:q")
	meta, _, err := l.CreatePractice(human, bot)
	if err != nil {
		t.Fatal(err)
	}
	if err := l.Delete(meta.ID); err != nil {
		t.Fatal(err)
	}
	after := series(t, metrics.Registry)
	for k, v := range after {
		if (strings.HasPrefix(k, "cmdctrl_games_") || strings.HasPrefix(k, "cmdctrl_game_duration")) && v != before[k] {
			t.Errorf("%s moved by %v for a practice table", k, v-before[k])
		}
	}
}

// The WebSocket counters: an upgrade with no session is refused as
// unauthorized; a seat's connect, frames and disconnect each count
// once.
func TestWSCountersMoveByOne(t *testing.T) {
	srv, l, hub, _ := metricsStack(t)
	meta, _ := l.Create("Sockets")
	token, _ := mustJoin(t, srv, meta, "Alice")

	names := []string{
		"cmdctrl_ws_upgrade_rejections_total{reason=unauthorized}",
		"cmdctrl_ws_connects_total{role=seat}",
		"cmdctrl_ws_disconnects_total{role=seat}",
		"cmdctrl_ws_frames_total{direction=out,type=snapshot}",
		"cmdctrl_ws_frames_total{direction=in,type=ping}",
		"cmdctrl_ws_frames_total{direction=out,type=pong}",
		"cmdctrl_ws_frames_total{direction=in,type=other}",
		"cmdctrl_ws_frames_total{direction=out,type=error}",
	}
	base := map[string]float64{}
	for _, n := range names {
		base[n] = event(t, n)
	}
	moved := func(n string) float64 { return event(t, n) - base[n] }

	if _, resp, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http")+"/ws", nil); err == nil {
		t.Fatal("an upgrade with no session was admitted")
	} else if resp != nil {
		resp.Body.Close()
	}
	if got := moved(names[0]); got != 1 {
		t.Errorf("unauthorized rejections moved by %v, want 1", got)
	}

	conn := mustDial(t, srv, token)
	if got := moved(names[1]); got != 1 {
		t.Errorf("seat connects moved by %v, want 1", got)
	}
	if got := moved(names[3]); got != 1 {
		t.Errorf("outbound snapshots moved by %v, want 1", got)
	}

	// A ping is one frame in and one out; a frame of no known kind is
	// one "other" in and one error out.
	for _, f := range []string{`{"v":0,"kind":"ping","id":"1"}`, `{"v":0,"kind":"teleport","id":"2"}`} {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(f)); err != nil {
			t.Fatal(err)
		}
		if _, _, err := conn.ReadMessage(); err != nil {
			t.Fatal(err)
		}
	}
	for _, n := range names[4:] {
		if got := moved(n); got != 1 {
			t.Errorf("%s moved by %v, want 1", n, got)
		}
	}

	_ = conn.Close()
	waitFor(t, func() bool { return moved(names[2]) >= 1 })
	if got := moved(names[2]); got != 1 || hub.Count() != 0 {
		t.Errorf("seat disconnects moved by %v (hub has %d), want 1 and 0", got, hub.Count())
	}
}

// ADR 0123 §2's locking rule under -race, and ADR 0124 §5's for
// Lobby.LiveTables: scrapes run while tables are
// created, joined, started, played, watched, archived and deleted and
// sockets come and go. A lock-order inversion between the collector
// and the hub, a room or the lobby would hang here; a data race fails
// the race detector.
func TestScrapeDuringTableTraffic(t *testing.T) {
	// Six tables' joins and spectates in a burst, from one address.
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	srv, l, _, reg := metricsStack(t)
	reg.MustRegister(metrics.NewUsersCollector(staticUsers{}, l, nil))

	stop := make(chan struct{})
	var scrapes sync.WaitGroup
	for range 3 {
		scrapes.Add(1)
		go func() {
			defer scrapes.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				if _, err := reg.Gather(); err != nil {
					t.Errorf("gather: %v", err)
					return
				}
				if _, err := metrics.Registry.Gather(); err != nil {
					t.Errorf("gather: %v", err)
					return
				}
				// The admin views' copy keeps the same rule (ADR 0124
				// §5): l.mu only, no room or game lock.
				_ = l.LiveTables()
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		var tables sync.WaitGroup
		for i := range 6 {
			tables.Add(1)
			go func() {
				defer tables.Done()
				meta, err := l.Create(fmt.Sprintf("Traffic %d", i))
				if err != nil {
					t.Error(err)
					return
				}
				token, alice, err := joinOverHTTP(srv, meta, "Alice")
				if err != nil {
					t.Error(err)
					return
				}
				_, bob, err := l.Join(meta.ID, meta.InviteToken, "Bob")
				if err != nil {
					t.Error(err)
					return
				}
				for _, p := range []uuid.UUID{alice, bob} {
					if _, err := l.SetDeck(meta.ID, p, "deck", botDeck(20)); err != nil {
						t.Error(err)
						return
					}
				}
				if _, err := l.Start(meta.ID); err != nil {
					t.Error(err)
					return
				}
				spec, err := spectateOverHTTP(srv, meta)
				if err != nil {
					t.Error(err)
					return
				}
				seat, err := dialAdmitted(srv, token)
				if err != nil {
					t.Error(err)
					return
				}
				defer seat.Close()
				watcher, err := dialAdmitted(srv, spec)
				if err != nil {
					t.Error(err)
					return
				}
				defer watcher.Close()
				_ = seat.WriteMessage(websocket.TextMessage, []byte(`{"v":0,"kind":"ping","id":"p"}`))
				l.List()
				switch i % 3 {
				case 0: // played to its end
					room := l.RoomOf(meta.ID)
					if _, _, err := room.Apply(bob, func() error { return room.Game.Concede(bob) }); err != nil {
						t.Error(err)
					}
				case 1: // closed by archiving
					if _, err := l.SetArchived(meta.ID, true); err != nil {
						t.Error(err)
					}
				}
				_ = seat.Close()
				_ = watcher.Close()
				if err := l.Delete(meta.ID); err != nil {
					t.Error(err)
				}
			}()
		}
		tables.Wait()
	}()

	select {
	case <-done:
	case <-time.After(60 * time.Second):
		t.Fatal("table traffic did not finish beside the scrapes: a lock-order inversion?")
	}
	close(stop)
	scrapes.Wait()
	if err := metrics.CheckClosedLabels(reg); err != nil {
		t.Fatalf("label rule broken:\n%v", err)
	}
	if err := metrics.CheckClosedLabels(metrics.Registry); err != nil {
		t.Fatalf("label rule broken:\n%v", err)
	}
}
