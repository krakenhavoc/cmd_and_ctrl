package lobby

// away_test.go: ADR 0177, the table moving on without a player who is
// gone: when a seat becomes movable, who may answer, and what each
// answer does. The presence is a fake, so the clock is the test's.

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// fakeAway is the hub's presence, set by the test. A seat it has not
// been told about is present.
type fakeAway struct {
	mu        sync.Mutex
	state     map[uuid.UUID]string
	since     map[uuid.UUID]time.Time
	refreshed int
}

func newFakeAway() *fakeAway {
	return &fakeAway{state: map[uuid.UUID]string{}, since: map[uuid.UUID]time.Time{}}
}

func (f *fakeAway) SeatPresence(_, seat uuid.UUID) (string, time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.state[seat], f.since[seat]
}

func (f *fakeAway) RefreshPresence(uuid.UUID) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.refreshed++
}

func (f *fakeAway) set(seat uuid.UUID, state string, since time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state[seat], f.since[seat] = state, since
}

func (f *fakeAway) refreshes() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.refreshed
}

// fakeStandIns is a bot host that can also stand in for a seat.
type fakeStandIns struct {
	*fakeBotHost
	mu      sync.Mutex
	running map[uuid.UUID]aiseat.SeatSpec
	stopped []uuid.UUID
}

func newFakeStandIns() *fakeStandIns {
	h := &fakeStandIns{fakeBotHost: newFakeBotHost(), running: map[uuid.UUID]aiseat.SeatSpec{}}
	h.tiers = []string{"random", StandInTier}
	return h
}

func (f *fakeStandIns) StartStandIn(_ *ws.Room, seat aiseat.SeatSpec) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.running[seat.PlayerID] = seat
}

func (f *fakeStandIns) StopStandIn(_, seat uuid.UUID) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.running, seat)
	f.stopped = append(f.stopped, seat)
}

func (f *fakeStandIns) runs(seat uuid.UUID) (aiseat.SeatSpec, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.running[seat]
	return s, ok
}

type awayStack struct {
	*seatRemoveStack
	away  *fakeAway
	bots  *fakeStandIns
	clock *time.Time
	meta  GameMeta
	// alice hosts; bob is the seat that goes away; carol is the third.
	alice, bob, carol sessionResponse
	start             time.Time
}

func newAwayStack(t *testing.T) *awayStack {
	t.Helper()
	s := &awayStack{seatRemoveStack: newSeatRemoveStack(t), away: newFakeAway(), bots: newFakeStandIns()}
	s.l.SetBotHost(s.bots)
	s.l.SetAwayPresence(s.away)
	s.start = time.Date(2026, 10, 14, 21, 0, 0, 0, time.UTC)
	now := s.start
	s.clock = &now
	s.l.SetAwayClock(func() time.Time { return *s.clock })
	meta, seats := s.startedTable(t, uuid.New(), "Ana", "Bo", "Cy")
	s.meta, s.alice, s.bob, s.carol = meta, seats[0], seats[1], seats[2]
	if len(legal.EnumerateFor(s.room().Game, s.bob.PlayerID)) == 0 {
		t.Fatal("the fixture's table is not waiting on Bo, so nothing here can be movable")
	}
	s.away.set(s.bob.PlayerID, protocol.PresenceOffline, s.start)
	return s
}

func (s *awayStack) room() *ws.Room { return s.l.RoomOf(s.meta.ID) }

func (s *awayStack) at(d time.Duration) {
	at := s.start.Add(d)
	*s.clock = at
}

func (s *awayStack) answer(t *testing.T, token string, seat uuid.UUID, action string) *http.Response {
	t.Helper()
	return postJSON(t, s.srv, "/games/"+s.meta.ID.String()+"/seats/"+seat.String()+"/away", token,
		map[string]string{"action": action})
}

func (s *awayStack) awayLog(t *testing.T) []string {
	t.Helper()
	var out []string
	for _, e := range protocol.ViewOfGameFor(s.room().Game, "").Log {
		if e.Kind == protocol.LogAwaySeat {
			out = append(out, e.Text)
		}
	}
	return out
}

func TestASeatBecomesMovableAfterTheAwayTime(t *testing.T) {
	s := newAwayStack(t)
	bo := s.bob.PlayerID

	s.at(3*time.Minute - time.Second)
	s.l.TickAway()
	if s.room().Movable(bo) {
		t.Fatal("movable before the away time")
	}
	before := s.away.refreshes()
	s.at(3 * time.Minute)
	s.l.TickAway()
	if !s.room().Movable(bo) {
		t.Fatal("not movable after 3 minutes offline while the table waits on the seat")
	}
	if s.away.refreshes() == before {
		t.Error("the table was not sent a frame when the seat became movable")
	}

	// Coming back withdraws it.
	s.away.set(bo, "", time.Time{})
	s.l.TickAway()
	if s.room().Movable(bo) {
		t.Error("still movable after the seat came back")
	}

	// A hidden tab never makes a seat movable.
	s.away.set(bo, protocol.PresenceHidden, s.start)
	s.at(30 * time.Minute)
	s.l.TickAway()
	if s.room().Movable(bo) {
		t.Error("a hidden seat is movable")
	}
}

func TestAwayAfterIsTheTablesSetting(t *testing.T) {
	s := newAwayStack(t)
	bo := s.bob.PlayerID
	one, never := 1, 0
	if err := s.room().Game.UpdateSettings(uuid.Nil, game.SettingsPatch{AwayAfter: &one}); err != nil {
		t.Fatal(err)
	}
	s.at(time.Minute)
	s.l.TickAway()
	if !s.room().Movable(bo) {
		t.Fatal("away_after 1: not movable after a minute")
	}
	if err := s.room().Game.UpdateSettings(uuid.Nil, game.SettingsPatch{AwayAfter: &never}); err != nil {
		t.Fatal(err)
	}
	s.at(time.Hour)
	s.l.TickAway()
	if s.room().Movable(bo) {
		t.Error("away_after 0: movable")
	}
}

func TestWhoMayAnswerTheAwayOffer(t *testing.T) {
	s := newAwayStack(t)
	bo := s.bob.PlayerID
	s.at(4 * time.Minute)
	s.l.TickAway()

	// The host is at the table: only the host (and the admin) decides.
	s.wantStatus(t, s.answer(t, "", bo, AwayActionWait), http.StatusUnauthorized)
	s.wantStatus(t, s.answer(t, s.carol.Token, bo, AwayActionWait), http.StatusForbidden)
	s.wantStatus(t, s.answer(t, s.bob.Token, bo, AwayActionWait), http.StatusForbidden)
	s.wantStatus(t, s.answer(t, s.alice.Token, bo, "kick"), http.StatusBadRequest)

	// The host is gone too: any present player decides.
	s.away.set(s.alice.PlayerID, protocol.PresenceOffline, s.start)
	s.wantStatus(t, s.answer(t, s.carol.Token, bo, AwayActionWait), http.StatusOK)
	// The answer took the offer away for everyone at once.
	if s.room().Movable(bo) {
		t.Error("still movable after someone chose to keep waiting")
	}
	s.wantStatus(t, s.answer(t, s.carol.Token, bo, AwayActionWait), http.StatusConflict)

	// An offline player is refused even when nobody else is there.
	s.away.set(s.carol.PlayerID, protocol.PresenceOffline, s.start)
	s.at(8 * time.Minute)
	s.wantStatus(t, s.answer(t, s.carol.Token, bo, AwayActionConcede), http.StatusForbidden)
	// The admin always may.
	s.wantStatus(t, s.answer(t, adminToken(t, s.srv), bo, AwayActionConcede), http.StatusOK)
}

func TestKeepWaitingHidesTheOfferForAnotherAwayTime(t *testing.T) {
	s := newAwayStack(t)
	bo := s.bob.PlayerID
	s.at(3 * time.Minute)
	s.l.TickAway()
	s.wantStatus(t, s.answer(t, s.alice.Token, bo, AwayActionWait), http.StatusOK)

	s.at(6*time.Minute - time.Second)
	s.l.TickAway()
	if s.room().Movable(bo) {
		t.Fatal("the offer came back before another away time")
	}
	s.at(6 * time.Minute)
	s.l.TickAway()
	if !s.room().Movable(bo) {
		t.Fatal("the offer did not come back after another away time")
	}
	if got := s.awayLog(t); len(got) != 0 {
		t.Errorf("keep waiting wrote %q to the log", got)
	}
}

func TestKeepWaitingIsPerAbsence(t *testing.T) {
	s := newAwayStack(t)
	bo := s.bob.PlayerID
	s.at(3 * time.Minute)
	s.l.TickAway()
	s.wantStatus(t, s.answer(t, s.alice.Token, bo, AwayActionWait), http.StatusOK)
	// Bo came back for a moment and left again at 4m: a new absence,
	// with its own clock, which the old "keep waiting" does not cover.
	s.away.set(bo, protocol.PresenceOffline, s.start.Add(4*time.Minute))
	s.at(7 * time.Minute)
	s.l.TickAway()
	if !s.room().Movable(bo) {
		t.Fatal("a new absence was held to the last one's keep waiting")
	}
}

func TestHandingASeatToABot(t *testing.T) {
	s := newAwayStack(t)
	bo := s.bob.PlayerID
	s.at(3*time.Minute + 30*time.Second)
	s.l.TickAway()
	s.wantStatus(t, s.answer(t, s.alice.Token, bo, AwayActionStandIn), http.StatusOK)

	spec, ok := s.bots.runs(bo)
	if !ok || spec.Tier != StandInTier {
		t.Fatalf("no %s runner on Bo's seat: %+v %v", StandInTier, spec, ok)
	}
	if !s.room().StandIn(bo) {
		t.Error("the room does not mark Bo's seat")
	}
	meta, _ := s.l.Get(s.meta.ID)
	seat, _ := findSeat(meta.Players, bo)
	if seat.StandIn != StandInTier || seat.StandInSince == nil || seat.IsBot {
		t.Errorf("seat = %+v; want a stand-in, not a bot", seat)
	}
	if p := s.room().Game.PlayerByID(bo); p.Eliminated || p.IsBot {
		t.Error("the hand-over changed the player in the engine")
	}
	if got := s.awayLog(t); len(got) != 1 || got[0] != "Ana handed Bo's seat to a bot (offline 3m)" {
		t.Errorf("log %q", got)
	}
	// One answer wins.
	s.wantStatus(t, s.answer(t, s.alice.Token, bo, AwayActionConcede), http.StatusConflict)
	s.l.TickAway()
	if s.room().Movable(bo) {
		t.Error("a seat a bot plays is movable")
	}

	// Only Bo takes it back, and only with the button.
	takeBack := func(token string) *http.Response {
		return postJSON(t, s.srv, "/games/"+s.meta.ID.String()+"/seats/me/take-back", token, nil)
	}
	s.wantStatus(t, takeBack(s.carol.Token), http.StatusConflict)
	resp := takeBack(s.bob.Token)
	s.wantStatus(t, resp, http.StatusOK)
	if _, ok := s.bots.runs(bo); ok {
		t.Error("the runner is still on Bo's seat")
	}
	if s.room().StandIn(bo) {
		t.Error("the room still marks Bo's seat")
	}
	if got := s.awayLog(t); len(got) != 2 || got[1] != "Bo took their seat back" {
		t.Errorf("log %q", got)
	}
	s.wantStatus(t, takeBack(s.bob.Token), http.StatusConflict)
}

func TestConcedingForAnAwayPlayer(t *testing.T) {
	s := newAwayStack(t)
	bo := s.bob.PlayerID
	s.at(5 * time.Minute)
	s.l.TickAway()
	resp := s.answer(t, s.alice.Token, bo, AwayActionConcede)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status %d: %s", resp.StatusCode, body)
	}
	var meta GameMeta
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		t.Fatal(err)
	}
	p := s.enginePlayer(t, s.meta.ID, bo)
	if !p.Eliminated || p.Removed {
		t.Errorf("Bo: eliminated=%v removed=%v, want a concession and no removal", p.Eliminated, p.Removed)
	}
	if s.room().Game.SeatRemoved(bo) {
		t.Error("Bo's session would be refused after a concession made for them")
	}
	if got := s.awayLog(t); len(got) != 1 || got[0] != "Ana conceded for Bo (offline 5m)" {
		t.Errorf("log %q", got)
	}
	s.wantStatus(t, s.answer(t, s.alice.Token, bo, AwayActionConcede), http.StatusConflict)
}

func TestAStandInComesBackAfterARestart(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	bots := newFakeStandIns()
	l.SetBotHost(bots)
	away := newFakeAway()
	l.SetAwayPresence(away)
	start := time.Date(2026, 10, 14, 21, 0, 0, 0, time.UTC)
	l.SetAwayClock(func() time.Time { return start.Add(10 * time.Minute) })

	meta, err := l.Create("Restart")
	if err != nil {
		t.Fatal(err)
	}
	var ids []uuid.UUID
	for _, n := range []string{"Ana", "Bo", "Cy"} {
		_, id, err := l.Join(meta.ID, meta.InviteToken, n)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := l.SetDeck(meta.ID, id, "deck", botDeck(30)); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatal(err)
	}
	bo := ids[1]
	away.set(bo, protocol.PresenceOffline, start)
	if _, err := l.AnswerAway(meta.ID, adminPrincipal(), true, bo, AwayActionStandIn); err != nil {
		t.Fatalf("hand-over: %v", err)
	}

	l2, _ := newDurableLobby(t, dir)
	bots2 := newFakeStandIns()
	l2.SetBotHost(bots2)
	if n := l2.RestoreFromDisk(quietLogger()); n != 1 {
		t.Fatalf("restored %d games, want 1", n)
	}
	if _, ok := bots2.runs(bo); !ok {
		t.Error("the restored table does not run Bo's stand-in")
	}
	if !l2.RoomOf(meta.ID).StandIn(bo) {
		t.Error("the restored room does not mark Bo's seat")
	}
	back, _ := l2.Get(meta.ID)
	if seat, _ := findSeat(back.Players, bo); seat.StandIn != StandInTier || seat.StandInSince == nil ||
		!seat.StandInSince.Equal(start.Add(10*time.Minute)) {
		t.Errorf("restored seat = %+v", seat)
	}
}

func adminPrincipal() auth.Principal {
	return auth.Principal{Role: auth.RoleAdmin, AdminID: uuid.New()}
}
