package lobby

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// botstats_test.go covers GET /games/{id}/bot/stats (#505 part 3).
//
// http.go now mounts BotStatsHandler on Handler(cfg)'s own mux (see
// its mux.Handle(BotStatsRoute, BotStatsHandler(c)) line), so most of
// these tests mount the handler standalone anyway — on its own bare
// mux, via newBotStatsServer — to exercise its own status-code and
// payload behaviour without the rest of Handler's routes in the way.
// TestBotStatsIsRoutedThroughTheRealMux below is the one that goes
// through Handler(cfg) itself, covering the thing that was actually
// broken before that line existed: a fully tested handler nothing
// ever dispatched to.

// fakeBotStatsHost wraps fakeBotHost (bots_test.go) with the one
// extra method BotStats asks for, so these tests get the existing
// BotHost recording behaviour (Tiers/StartBots/StopBots) for free and
// only have to fake Runners.
type fakeBotStatsHost struct {
	*fakeBotHost
	runners map[uuid.UUID][]*aiseat.Runner
}

func newFakeBotStatsHost() *fakeBotStatsHost {
	return &fakeBotStatsHost{fakeBotHost: newFakeBotHost(), runners: map[uuid.UUID][]*aiseat.Runner{}}
}

func (f *fakeBotStatsHost) Runners(gameID uuid.UUID) []*aiseat.Runner {
	return f.runners[gameID]
}

// botStatsFunnelPolicy is a minimal Policy that also reports a fixed
// PolicyStats, so the endpoint's funnel-projection field can be
// exercised without a live model funnel.
type botStatsFunnelPolicy struct {
	stats aiseat.PolicyStats
}

func (botStatsFunnelPolicy) Name() string { return "assisted" }

func (botStatsFunnelPolicy) Decide(_ context.Context, in aiseat.Input) (aiseat.Decision, error) {
	if len(in.Moves) == 0 {
		return aiseat.Decision{}, aiseat.ErrNoMoves
	}
	return aiseat.Decision{Index: 0, Reason: "fake"}, nil
}

func (p botStatsFunnelPolicy) PolicyStats() aiseat.PolicyStats { return p.stats }

// newBotStatsServer mounts BotStatsHandler alone, on its own route —
// see the file comment.
func newBotStatsServer(t *testing.T, cfg Config) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.Handle(BotStatsRoute, BotStatsHandler(cfg))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// newBotStatsTestRoom builds a minimal two-seat active game to start
// a real *aiseat.Runner against, the same shape cmd/server's
// seedDemoGame uses for the same reason: PolicyStats and Stats read
// off the runner, not off a mock.
func newBotStatsTestRoom(t *testing.T) *ws.Room {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < 2; i++ {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Commander %d", i), uuid.Nil)}
		for j := 0; j < 40; j++ {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d-%d", i, j), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(nil); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return ws.NewRoom(g, discardLogger(), "")
}

// TestBotStatsIsRoutedThroughTheRealMux is #505's actual regression
// target: BotStatsHandler existed, fully tested, and was reachable
// only when a test (or nothing) mounted it on a standalone mux by
// hand — see the other tests in this file, which predate http.go's
// mux.Handle(BotStatsRoute, BotStatsHandler(c)) line and deliberately
// exercise the handler that way. This one goes through Handler(cfg)
// itself, the same constructor every other route in this package is
// tested through, so a regression that unroutes the pattern again
// (or loosens its admin gate) fails here rather than only in a unit
// test of the handler in isolation.
func TestBotStatsIsRoutedThroughTheRealMux(t *testing.T) {
	l := newTestLobby(t)
	host := newFakeBotStatsHost()
	l.SetBotHost(host)
	a := newTestAuth()
	meta, _ := l.Create("FNM")
	srv := httptest.NewServer(Handler(Config{Lobby: l, Auth: a, Bots: host, Log: discardLogger()}))
	t.Cleanup(srv.Close)

	// Reachable and admin-gated: an admin session gets the real
	// response shape through the real mux.
	token := adminSession(t, a)
	resp := doGet(t, srv, "/games/"+meta.ID.String()+"/bot/stats", token)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		t.Fatalf("admin through real mux: got %d, want 200 (body=%s)", resp.StatusCode, body)
	}
	var out botStatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if len(out.Seats) != 0 {
		t.Errorf("seats: %+v, want none", out.Seats)
	}

	// A seated player is not an admin — the real mux's route must
	// carry the same auth.RoleAdmin gate BotStatsHandler wraps itself
	// in, not a looser one some other registration point might apply.
	playerTok := playerSession(t, a, meta.ID)
	resp = doGet(t, srv, "/games/"+meta.ID.String()+"/bot/stats", playerTok)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("player through real mux: got %d, want %d (body=%s)", resp.StatusCode, http.StatusForbidden, body)
	}
}

func TestBotStatsRequiresAdmin(t *testing.T) {
	l := newTestLobby(t)
	host := newFakeBotStatsHost()
	l.SetBotHost(host)
	a := newTestAuth()
	meta, _ := l.Create("FNM")
	srv := newBotStatsServer(t, Config{Lobby: l, Auth: a, Bots: host})

	// A seated player's session is not enough.
	playerTok := playerSession(t, a, meta.ID)
	resp := doGet(t, srv, "/games/"+meta.ID.String()+"/bot/stats", playerTok)
	if resp.StatusCode != http.StatusForbidden {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("player: got %d, want %d (body=%s)", resp.StatusCode, http.StatusForbidden, body)
	}
	resp.Body.Close()

	// No credential at all.
	resp = doGet(t, srv, "/games/"+meta.ID.String()+"/bot/stats", "")
	if resp.StatusCode != http.StatusUnauthorized {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("anonymous: got %d, want %d (body=%s)", resp.StatusCode, http.StatusUnauthorized, body)
	}
	resp.Body.Close()
}

func TestBotStatsReturns404ForAnUnknownGame(t *testing.T) {
	l := newTestLobby(t)
	host := newFakeBotStatsHost()
	l.SetBotHost(host)
	a := newTestAuth()
	srv := newBotStatsServer(t, Config{Lobby: l, Auth: a, Bots: host})
	token := adminSession(t, a)

	resp := doGet(t, srv, "/games/"+uuid.New().String()+"/bot/stats", token)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("unknown game: got %d, want 404 (body=%s)", resp.StatusCode, body)
	}
}

func TestBotStatsReturns503WithoutABotHost(t *testing.T) {
	l := newTestLobby(t)
	a := newTestAuth()
	meta, _ := l.Create("FNM")
	srv := newBotStatsServer(t, Config{Lobby: l, Auth: a})
	token := adminSession(t, a)

	resp := doGet(t, srv, "/games/"+meta.ID.String()+"/bot/stats", token)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("no host: got %d, want 503 (body=%s)", resp.StatusCode, body)
	}
}

func TestBotStatsReturns503WhenTheHostCannotReportRunners(t *testing.T) {
	// A BotHost that only implements the narrower BotHost interface
	// (Tiers/StartBots/StopBots) is a supported shape — every bot
	// test double in this package builds one — and must not panic.
	l := newTestLobby(t)
	host := newFakeBotHost()
	l.SetBotHost(host)
	a := newTestAuth()
	meta, _ := l.Create("FNM")
	srv := newBotStatsServer(t, Config{Lobby: l, Auth: a, Bots: host})
	token := adminSession(t, a)

	resp := doGet(t, srv, "/games/"+meta.ID.String()+"/bot/stats", token)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("narrower host: got %d, want 503 (body=%s)", resp.StatusCode, body)
	}
}

func TestBotStatsReturnsEmptyListForAGameWithNoBotSeats(t *testing.T) {
	l := newTestLobby(t)
	host := newFakeBotStatsHost()
	l.SetBotHost(host)
	a := newTestAuth()
	meta, _ := l.Create("FNM")
	srv := newBotStatsServer(t, Config{Lobby: l, Auth: a, Bots: host})
	token := adminSession(t, a)

	resp := doGet(t, srv, "/games/"+meta.ID.String()+"/bot/stats", token)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("bot stats: got %d, want 200 (body=%s)", resp.StatusCode, body)
	}
	var out botStatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if len(out.Seats) != 0 {
		t.Errorf("seats: %+v, want none", out.Seats)
	}
}

// One object per bot seat, carrying Runner.Stats() — including #505
// part 1's latency percentiles — and, when the policy has a model
// funnel underneath it, Runner.PolicyStats() too.
func TestBotStatsReturnsOneObjectPerSeat(t *testing.T) {
	l := newTestLobby(t)
	host := newFakeBotStatsHost()
	l.SetBotHost(host)
	a := newTestAuth()
	meta, _ := l.Create("FNM")
	srv := newBotStatsServer(t, Config{Lobby: l, Auth: a, Bots: host})
	token := adminSession(t, a)

	room := newBotStatsTestRoom(t)
	randomSeat := room.Game.Seats[0].ID
	funnelSeat := room.Game.Seats[1].ID
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// A seat with no model underneath it: PolicyStats must be absent.
	r0 := aiseat.Start(ctx, room, randomSeat, aiseat.NewRandomPolicy(nil), aiseat.Config{MinThink: time.Hour}, nil, discardLogger())
	// A seat whose policy funnels through a model: PolicyStats must
	// carry that funnel's own counters, verbatim.
	want := aiseat.PolicyStats{Windows: 4, ModelCalls: 2, ByLayer: map[string]int64{"C": 2}}
	r1 := aiseat.Start(ctx, room, funnelSeat, botStatsFunnelPolicy{stats: want}, aiseat.Config{MinThink: time.Hour}, nil, discardLogger())
	t.Cleanup(func() {
		cancel()
		<-r0.Done()
		<-r1.Done()
	})
	host.runners[meta.ID] = []*aiseat.Runner{r0, r1}

	resp := doGet(t, srv, "/games/"+meta.ID.String()+"/bot/stats", token)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("bot stats: got %d, want 200 (body=%s)", resp.StatusCode, body)
	}
	var out botStatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if len(out.Seats) != 2 {
		t.Fatalf("seats: %+v", out.Seats)
	}

	random := out.Seats[0]
	if random.Seat != randomSeat {
		t.Errorf("seat[0] = %s, want %s", random.Seat, randomSeat)
	}
	if random.Policy != "random" {
		t.Errorf("seat[0] policy = %q, want %q", random.Policy, "random")
	}
	if random.PolicyStats != nil {
		t.Errorf("a random seat reported PolicyStats: %+v", random.PolicyStats)
	}

	funnel := out.Seats[1]
	if funnel.Seat != funnelSeat {
		t.Errorf("seat[1] = %s, want %s", funnel.Seat, funnelSeat)
	}
	if funnel.Policy != "assisted" {
		t.Errorf("seat[1] policy = %q, want %q", funnel.Policy, "assisted")
	}
	if funnel.PolicyStats == nil {
		t.Fatal("a funnel seat reported no PolicyStats")
	}
	if got := *funnel.PolicyStats; got.Windows != want.Windows || got.ModelCalls != want.ModelCalls || got.ByLayer["C"] != want.ByLayer["C"] {
		t.Errorf("PolicyStats = %+v, want %+v", got, want)
	}
}
