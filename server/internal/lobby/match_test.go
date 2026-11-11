package lobby

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/gamestats"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// match_test.go: a match through the lobby (ADR 0222 §3, #5436): the
// match length on a new table, a best-of-three as one finished table,
// the between-games timer, and a restart in the middle of sideboarding.

func TestCreateABestOfThreeTable(t *testing.T) {
	s := newFormatStack(t)
	code, out, _, raw := s.create(t, s.admin(t), createGameRequest{Name: "Bo3", Format: "standard", BestOf: 3})
	if code != http.StatusCreated {
		t.Fatalf("create: %d %s", code, raw)
	}
	if out.BestOf != 3 || !strings.Contains(raw, `"best_of":3`) {
		t.Errorf("created meta: best_of %d; body %s", out.BestOf, raw)
	}
	g, err := s.lobby.LookupGame(out.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got := g.TableSettingsSnapshot().EffectiveBestOf(); got != 3 {
		t.Errorf("the table is best-of-%d, want 3", got)
	}
	if listed, err := s.lobby.Get(out.ID); err != nil || listed.BestOf != 3 {
		t.Errorf("listed meta: best_of %d, err %v", listed.BestOf, err)
	}

	// A best-of-one Standard table and a Commander table name no match
	// length at all.
	for _, body := range []createGameRequest{{Name: "Bo1", Format: "standard"}, {Name: "EDH"}, {Name: "EDH one", BestOf: 1}} {
		code, out, _, raw := s.create(t, s.admin(t), body)
		if code != http.StatusCreated {
			t.Fatalf("create %q: %d %s", body.Name, code, raw)
		}
		if out.BestOf != 0 || strings.Contains(raw, "best_of") {
			t.Errorf("%q: best_of %d; body %s", body.Name, out.BestOf, raw)
		}
	}

	// A Commander table has no match, and two is not a match length.
	before := len(s.lobby.List())
	for _, body := range []createGameRequest{{Name: "EDH Bo3", BestOf: 3}, {Name: "Bo2", Format: "standard", BestOf: 2}} {
		if code, _, _, raw := s.create(t, s.admin(t), body); code != http.StatusBadRequest {
			t.Errorf("create %q: %d %s, want 400", body.Name, code, raw)
		}
	}
	if after := len(s.lobby.List()); after != before {
		t.Errorf("a refused create left a table behind: %d tables, was %d", after, before)
	}
}

// matchCards is a sixty-card deck and a fifteen-card sideboard.
func matchCards(owner string) []game.Card {
	deck := make([]game.Card, 0, 75)
	for i := 0; i < 60; i++ {
		deck = append(deck, game.NewCard(fmt.Sprintf("%s Main %d", owner, i+1), uuid.Nil))
	}
	for i := 0; i < 15; i++ {
		c := game.NewCard(fmt.Sprintf("%s Side %d", owner, i+1), uuid.Nil)
		c.IsSideboard = true
		deck = append(deck, c)
	}
	return deck
}

// startMatch is a started best-of-three Standard table: Alice, then Bob.
func startMatch(t *testing.T, l *Lobby, name string) (GameMeta, uuid.UUID, uuid.UUID) {
	t.Helper()
	meta, err := l.Create(name)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	std, three := game.FormatStandard, 3
	if _, err := l.UpdateSettings(meta.ID, uuid.Nil, game.SettingsPatch{Format: &std, BestOf: &three}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	_, alice, err := l.Join(meta.ID, meta.InviteToken, "Alice")
	if err != nil {
		t.Fatalf("Join Alice: %v", err)
	}
	_, bob, err := l.Join(meta.ID, meta.InviteToken, "Bob")
	if err != nil {
		t.Fatalf("Join Bob: %v", err)
	}
	for _, seat := range []struct {
		id   uuid.UUID
		name string
	}{{alice, "Alice"}, {bob, "Bob"}} {
		if _, err := l.SetDeck(meta.ID, seat.id, seat.name+"'s", matchCards(seat.name)); err != nil {
			t.Fatalf("SetDeck %s: %v", seat.name, err)
		}
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return meta, alice, bob
}

func concedeGame(t *testing.T, room *ws.Room, player uuid.UUID) {
	t.Helper()
	if _, _, err := room.Apply(player, func() error { return room.Game.Concede(player) }); err != nil {
		t.Fatalf("conceding: %v", err)
	}
}

func readyBoth(t *testing.T, room *ws.Room, players ...uuid.UUID) {
	t.Helper()
	for _, p := range players {
		if _, _, err := room.ApplyExternal(func() error { return room.Game.SetSideboardReady(p, true) }); err != nil {
			t.Fatalf("readying up: %v", err)
		}
	}
}

func listed(t *testing.T, l *Lobby, id uuid.UUID) GameMeta {
	t.Helper()
	for _, m := range l.List() {
		if m.ID == id {
			return m
		}
	}
	t.Fatalf("table %s is not listed", id)
	return GameMeta{}
}

func matchLine(m GameMeta) string {
	wins := []string{}
	for _, p := range m.Players {
		if p.MatchWins == nil {
			wins = append(wins, "-")
		} else {
			wins = append(wins, fmt.Sprint(*p.MatchWins))
		}
	}
	return fmt.Sprintf("best of %d, game %d, %s", m.BestOf, m.MatchGame, strings.Join(wins, "-"))
}

// countedStats is the SQL store, counting the records written to it.
type countedStats struct {
	*gamestats.SQLStore
	puts atomic.Int32
}

func (c *countedStats) Put(ctx context.Context, g gamestats.Game) error {
	c.puts.Add(1)
	return c.SQLStore.Put(ctx, g)
}

// A best-of-three is one finished table: it stays in play between its
// games, ends once, with the match's winner, and gets one statistics
// record that keeps each game.
func TestAMatchIsOneFinishedTable(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	d := openTestDB(t, dir)
	stats := &countedStats{SQLStore: gamestats.NewSQLStore(d)}
	l.SetGameStats(stats)
	meta, alice, bob := startMatch(t, l, "Match")
	room := l.RoomOf(meta.ID)

	if got := matchLine(listed(t, l, meta.ID)); got != "best of 3, game 1, 0-0" {
		t.Errorf("listed at the start: %s", got)
	}

	// Game 1 to Alice. The table is between games and still in play.
	concedeGame(t, room, bob)
	if !room.Game.SideboardingOpen() {
		t.Fatal("no sideboarding window after game 1")
	}
	m := listed(t, l, meta.ID)
	if m.State != "active" {
		t.Errorf("listed state between games = %q, want active", m.State)
	}
	if got := matchLine(m); got != "best of 3, game 1, 1-0" {
		t.Errorf("listed between games: %s", got)
	}
	if got, err := l.Get(meta.ID); err != nil || got.State != "active" {
		t.Errorf("meta between games: state %q, err %v", got.State, err)
	}

	// Game 2 to Bob, game 3 to Alice.
	readyBoth(t, room, alice, bob)
	if got := matchLine(listed(t, l, meta.ID)); got != "best of 3, game 2, 1-0" {
		t.Errorf("listed in game 2: %s", got)
	}
	concedeGame(t, room, alice)
	readyBoth(t, room, alice, bob)
	if n := stats.puts.Load(); n != 0 {
		t.Fatalf("%d statistics records written before the match ended", n)
	}
	concedeGame(t, room, bob)

	var rec gamestats.Game
	waitFor(t, func() bool {
		got, err := stats.Get(context.Background(), meta.ID)
		rec = got
		return err == nil
	})
	l.WaitGameStats()
	if n := stats.puts.Load(); n != 1 {
		t.Errorf("%d statistics records for one match, want 1", n)
	}

	ended, err := l.Get(meta.ID)
	if err != nil {
		t.Fatal(err)
	}
	if ended.State != "ended" {
		t.Errorf("state after the match = %q, want ended", ended.State)
	}
	if got := matchLine(listed(t, l, meta.ID)); got != "best of 3, game 3, 2-1" {
		t.Errorf("listed after the match: %s", got)
	}
	var winner *int
	var outcome string
	if err := d.QueryRow(`SELECT winner_seat, outcome FROM games WHERE id = ?`, meta.ID.String()).Scan(&winner, &outcome); err != nil {
		t.Fatal(err)
	}
	if winner == nil || *winner != 0 || outcome != game.OutcomeWin {
		t.Errorf("games row: winner_seat %v, outcome %q; want Alice's seat and a win", winner, outcome)
	}

	// One result per seat, the match's.
	if len(rec.Seats) != 2 {
		t.Fatalf("%d seat lines, want 2", len(rec.Seats))
	}
	for _, s := range rec.Seats {
		want := gamestats.ResultLoss
		if room.Game.Seats[s.Seat].ID == alice {
			want = gamestats.ResultWin
		}
		if s.Result != want {
			t.Errorf("seat %d: result %q, want %q", s.Seat, s.Result, want)
		}
	}
	// And each game kept.
	rows, err := d.Query(`SELECT game_no, outcome, winner_seat FROM match_games WHERE game_id = ? ORDER BY game_no`, meta.ID.String())
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var got []string
	for rows.Next() {
		var no int
		var out string
		var seat *int
		if err := rows.Scan(&no, &out, &seat); err != nil {
			t.Fatal(err)
		}
		got = append(got, fmt.Sprintf("%d:%s:%d", no, out, *seat))
	}
	if want := "1:win:0 2:win:1 3:win:0"; strings.Join(got, " ") != want {
		t.Errorf("match_games = %v, want %s", got, want)
	}
}

// A Commander table through the same lobby: no match on its meta, no
// between-games step, and no match_games rows.
func TestACommanderTableHasNoMatchInTheLobby(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	d := openTestDB(t, dir)
	stats := gamestats.NewSQLStore(d)
	l.SetGameStats(stats)
	meta, alice, _ := startTwoSeatGame(t, l, "EDH")
	room := l.RoomOf(meta.ID)
	if got := matchLine(listed(t, l, meta.ID)); got != "best of 0, game 0, ---" {
		t.Errorf("a Commander table is listed with a match: %s", got)
	}
	concedeGame(t, room, alice)
	if room.Game.SideboardingOpen() {
		t.Fatal("a Commander table opened a sideboarding window")
	}
	waitFor(t, func() bool {
		_, err := stats.Get(context.Background(), meta.ID)
		return err == nil
	})
	l.WaitGameStats()
	if got, err := l.Get(meta.ID); err != nil || got.State != "ended" || got.BestOf != 0 {
		t.Errorf("meta: state %q, best_of %d, err %v", got.State, got.BestOf, err)
	}
	var n int
	if err := d.QueryRow(`SELECT COUNT(*) FROM match_games`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("%d match_games rows for a Commander table", n)
	}
}

// The timer ends sideboarding: the lobby begins the next game at the
// deadline with nobody ready.
func TestTheTimerBeginsTheNextGame(t *testing.T) {
	l, _ := newDurableLobby(t, t.TempDir())
	meta, _, bob := startMatch(t, l, "Timer")
	room := l.RoomOf(meta.ID)
	concedeGame(t, room, bob)
	if n, between := room.Game.MatchMark(); n != 1 || !between {
		t.Fatalf("after game 1: game %d, between %v", n, between)
	}
	// Still waiting: the window is three minutes long.
	time.Sleep(50 * time.Millisecond)
	if !room.Game.SideboardingOpen() {
		t.Fatal("the next game began with the deadline minutes away")
	}
	// The commit wakes the watcher, which reads the new deadline.
	if _, _, err := room.ApplyExternal(func() error {
		room.Game.SetSideboardingDeadlineForTest(time.Now().Add(20 * time.Millisecond))
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	// Monotone: the game number only climbs.
	waitFor(t, func() bool { return room.Game.MatchGameNumber() >= 2 })
	if room.Game.SideboardingOpen() {
		t.Error("the window is still open in game 2")
	}
	if got, err := l.Get(meta.ID); err != nil || got.State != "active" {
		t.Errorf("meta in game 2: state %q, err %v", got.State, err)
	}
}

// A restart in the middle of sideboarding: the table comes back between
// games with its score, gives its players time to return, and plays on.
func TestAMatchSurvivesARestartDuringSideboarding(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	meta, alice, bob := startMatch(t, l, "Restart")
	room := l.RoomOf(meta.ID)
	concedeGame(t, room, bob)
	// The deadline passes while the server is down.
	if _, _, err := room.ApplyExternal(func() error {
		if err := room.Game.SetSideboardReady(alice, true); err != nil {
			return err
		}
		room.Game.SetSideboardingDeadlineForTest(time.Now().Add(-time.Hour))
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	_ = l

	l2, _ := newDurableLobby(t, dir)
	if n := l2.RestoreFromDisk(quietLogger()); n != 1 {
		t.Fatalf("restored %d games, want 1", n)
	}
	room2 := l2.RoomOf(meta.ID)
	if room2 == nil {
		t.Fatal("no room after the restart")
	}
	g := room2.Game
	if n, between := g.MatchMark(); n != 1 || !between {
		t.Fatalf("restored: game %d, between %v; want between games after game 1", n, between)
	}
	if got := matchLine(listed(t, l2, meta.ID)); got != "best of 3, game 1, 1-0" {
		t.Errorf("listed after the restart: %s", got)
	}
	if got, err := l2.Get(meta.ID); err != nil || got.State != "active" {
		t.Errorf("meta after the restart: state %q, err %v", got.State, err)
	}
	// It did not start game 2 on the deadline that passed.
	time.Sleep(100 * time.Millisecond)
	if deadline, open := g.SideboardingDeadline(); !open || time.Until(deadline) < 30*time.Second {
		t.Fatalf("after the restart: open %v, %v left; want the window open with time to come back", open, time.Until(deadline))
	}
	// Alice was ready before the restart; Bob readies up and game 2 begins.
	readyBoth(t, room2, bob)
	if n, between := g.MatchMark(); n != 2 || between {
		t.Fatalf("after Bob readied up: game %d, between %v", n, between)
	}
	if m := g.MatchSnapshot(); m.Wins(alice) != 1 || m.Wins(bob) != 0 {
		t.Errorf("score after the restart: %d-%d, want 1-0", m.Wins(alice), m.Wins(bob))
	}
}
