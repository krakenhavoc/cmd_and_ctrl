package lobby

// idle_test.go: ADR 0211, closing idle tables. The sweep's clock and
// every room's clock are the test's, so a day passes in a line.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/capacity"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics/metricstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// fakeEvictor (practice_test.go) records the closed tables' evictions.
type idleStack struct {
	l     *Lobby
	mgr   *ws.RoomManager
	bots  *fakeBotHost
	away  *fakeAway
	ev    *fakeEvictor
	start time.Time
	mu    sync.Mutex
	now   time.Time
}

// newIdleStack is a lobby with the shipped idle thresholds (24 h, 48 h,
// 1 h) and a test clock. dir is the data dir, "" for none.
func newIdleStack(t *testing.T, dir string, lim capacity.Limits) *idleStack {
	t.Helper()
	s := &idleStack{mgr: ws.NewRoomManager(quietLogger(), dir), bots: newFakeBotHost(), away: newFakeAway(), ev: &fakeEvictor{}}
	if dir == "" {
		s.l = NewLobby(s.mgr)
	} else {
		s.l = NewLobbyWithStore(s.mgr, NewSQLStore(openTestDB(t, dir)))
	}
	// The engine stamps created_at with the wall clock, so the test's
	// clock starts there.
	s.start = time.Now().UTC().Truncate(time.Millisecond)
	s.now = s.start
	s.l.SetIdleClock(s.clock)
	s.l.SetCapacity(lim, nil)
	s.l.SetBotHost(s.bots)
	s.l.SetAwayPresence(s.away)
	s.l.SetEvictor(s.ev)
	return s
}

// idleLimits are the shipped limits: the caps and weights, so a table's
// cost can be seen to fall, and the three idle thresholds.
func idleLimits() capacity.Limits { return capacity.Defaults() }

func (s *idleStack) clock() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.now
}

// at moves the clock to d after the start and runs the sweep.
func (s *idleStack) at(d time.Duration) {
	s.mu.Lock()
	s.now = s.start.Add(d)
	s.mu.Unlock()
	s.l.TickIdle()
}

// room hands the table's room the test's clock, so its moves are
// stamped on it.
func (s *idleStack) room(t *testing.T, id uuid.UUID) *ws.Room {
	t.Helper()
	r := s.l.RoomOf(id)
	if r == nil {
		t.Fatalf("no room for %s", id)
	}
	r.SetClock(s.clock)
	return r
}

// move is a person's own move at seat, at the current time.
func (s *idleStack) move(t *testing.T, id, seat uuid.UUID) {
	t.Helper()
	if _, _, err := s.room(t, id).ApplyAction(ws.ActionTag{Actor: seat}, func() error { return nil }); err != nil {
		t.Fatalf("move: %v", err)
	}
}

func (s *idleStack) entry(t *testing.T, id uuid.UUID) *gameEntry {
	t.Helper()
	s.l.mu.Lock()
	defer s.l.mu.Unlock()
	e, ok := s.l.games[id]
	if !ok {
		t.Fatalf("table %s is not resident", id)
	}
	return e
}

func (s *idleStack) archived(t *testing.T, id uuid.UUID) bool {
	t.Helper()
	return s.l.IsArchived(id)
}

// idleLines are the table's EventIdleClose labels, oldest first.
func idleLines(t *testing.T, room *ws.Room) []game.Event {
	t.Helper()
	var out []game.Event
	room.Game.ReadSnapshot(func() {
		for _, ev := range room.Game.Events {
			if ev.Kind == game.EventIdleClose {
				out = append(out, ev)
			}
		}
	})
	return out
}

func TestIdleLobbyTableClosesAfterADay(t *testing.T) {
	s := newIdleStack(t, "", idleLimits())
	creator := uuid.New()
	meta, err := s.l.CreateBy("Nobody came", creator)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.l.Join(meta.ID, meta.InviteToken, "Alice"); err != nil {
		t.Fatal(err)
	}
	if got := len(s.l.OpenTablesCreatedBy(creator)); got != 1 {
		t.Fatalf("open tables by the creator: %d, want 1", got)
	}
	if s.l.CapacityUsage().Load == 0 {
		t.Fatal("an open table costs nothing; the test cannot see it freed")
	}

	s.at(24*time.Hour - 5*time.Minute)
	if s.archived(t, meta.ID) {
		t.Fatal("closed at 23 h 55")
	}
	s.at(24*time.Hour + time.Minute)
	if !s.archived(t, meta.ID) {
		t.Fatal("still open after 24 h with no activity")
	}
	e := s.entry(t, meta.ID)
	if e.closedReason != ClosedIdleLobby {
		t.Errorf("closed_reason %q, want %q", e.closedReason, ClosedIdleLobby)
	}
	if lines := idleLines(t, e.room); len(lines) != 1 || lines[0].Label != game.IdleCloseLobby || lines[0].Amount != 24 {
		t.Errorf("log lines %+v, want one idle_lobby line of 24 hours", lines)
	}
	if got := s.l.CapacityUsage().Load; got != 0 {
		t.Errorf("a closed table still costs %d", got)
	}
	if got := len(s.l.OpenTablesCreatedBy(creator)); got != 0 {
		t.Errorf("the creator's slot is still taken: %d open", got)
	}
	if !s.ev.has(meta.ID) {
		t.Error("the closed table's sockets were not evicted")
	}
	rec, _, err := s.l.store.LoadGame(context.Background(), meta.ID)
	if err != nil {
		t.Fatal(err)
	}
	if rec.ClosedReason != ClosedIdleLobby || rec.ArchivedAt == nil || rec.Outcome != "" || rec.WinnerSeat != nil {
		t.Errorf("row %+v: want archived, idle_lobby, no outcome, no winner", rec)
	}
	if rec.LastActiveAt == nil {
		t.Error("the sweep never wrote last_active_at")
	}
}

func TestIdleLobbyTouchResetsTheClock(t *testing.T) {
	s := newIdleStack(t, "", idleLimits())
	meta, _ := s.l.Create("Slow to fill")
	_, alice, _ := s.l.Join(meta.ID, meta.InviteToken, "Alice")

	s.at(20 * time.Hour)
	if _, err := s.l.SetDeck(meta.ID, alice, "Alice's", botDeck(20)); err != nil {
		t.Fatal(err)
	}
	s.at(24*time.Hour + time.Minute)
	if s.archived(t, meta.ID) {
		t.Fatal("closed 4 hours after a seat set its deck")
	}
	s.at(44*time.Hour + time.Minute)
	if !s.archived(t, meta.ID) {
		t.Fatal("still open 24 hours after the last touch")
	}
}

func TestIdleLobbyAReadDoesNotTouch(t *testing.T) {
	s := newIdleStack(t, "", idleLimits())
	meta, _ := s.l.Create("Watched")
	_, _, _ = s.l.Join(meta.ID, meta.InviteToken, "Alice")
	s.at(23 * time.Hour)
	// A spectator, a preview and a read are not activity.
	if _, err := s.l.Spectate(meta.ID, meta.SpectatorInvite); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.l.Preview(meta.ID, meta.InviteToken); err != nil {
		t.Fatal(err)
	}
	if _, err := s.l.Get(meta.ID); err != nil {
		t.Fatal(err)
	}
	s.at(24*time.Hour + time.Minute)
	if !s.archived(t, meta.ID) {
		t.Fatal("a spectator, a preview or a read kept the table open")
	}
}

func TestIdleGameIsWarnedThenClosedAfterTwoDays(t *testing.T) {
	s := newIdleStack(t, "", idleLimits())
	meta, alice, _ := startTwoSeatGame(t, s.l, "Walked away")
	before := metricstest.Value(t, "cmdctrl_games_ended_total", metricstest.L{"outcome": "closed"})

	s.at(35 * time.Hour)
	if lines := idleLines(t, s.room(t, meta.ID)); len(lines) != 0 {
		t.Fatalf("warned at 35 h: %+v", lines)
	}
	s.at(36*time.Hour + time.Minute)
	lines := idleLines(t, s.room(t, meta.ID))
	if len(lines) != 1 || lines[0].Label != game.IdleCloseWarning || lines[0].Amount != 36 {
		t.Fatalf("at 36 h the log has %+v, want one 36-hour warning", lines)
	}
	closes, err := time.Parse(time.RFC3339, lines[0].Call)
	if err != nil {
		t.Fatalf("the warning's closing time %q: %v", lines[0].Call, err)
	}
	if d := closes.Sub(s.start.Add(48 * time.Hour)); d < -2*time.Second || d > 2*time.Second {
		t.Errorf("the warning says it closes at %v, want 48 h after the start (%v)", closes, s.start.Add(48*time.Hour))
	}
	view, _, err := s.room(t, meta.ID).Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if !logHas(view.Log, "No one has moved for 36 hours. This game closes at") {
		t.Errorf("the log does not read as the warning: %+v", view.Log)
	}

	s.at(47 * time.Hour)
	if s.archived(t, meta.ID) {
		t.Fatal("closed before 48 h")
	}
	s.at(48*time.Hour + time.Minute)
	if !s.archived(t, meta.ID) {
		t.Fatal("still open after 48 h with no move")
	}
	e := s.entry(t, meta.ID)
	if e.closedReason != ClosedIdleGame || e.outcome != "" || e.winnerSeat != nil {
		t.Errorf("closed with reason %q outcome %q winner %v; want idle_game and no result", e.closedReason, e.outcome, e.winnerSeat)
	}
	if got := metricstest.Value(t, "cmdctrl_games_ended_total", metricstest.L{"outcome": "closed"}); got != before+1 {
		t.Errorf("closed ends counted %v, want %v", got, before+1)
	}
	s.bots.mu.Lock()
	stopped := append([]uuid.UUID(nil), s.bots.stopped...)
	s.bots.mu.Unlock()
	if len(stopped) == 0 || stopped[len(stopped)-1] != meta.ID {
		t.Error("the closed game's bots were not stopped")
	}
	// A move after the close cannot reopen it.
	s.move(t, meta.ID, alice)
	s.at(49 * time.Hour)
	if !s.archived(t, meta.ID) {
		t.Error("a closed game came back open")
	}
}

func logHas(entries []protocol.LogEvent, want string) bool {
	for _, e := range entries {
		if strings.Contains(e.Text, want) {
			return true
		}
	}
	return false
}

func TestIdleGameAMoveClearsTheWarning(t *testing.T) {
	s := newIdleStack(t, "", idleLimits())
	meta, alice, _ := startTwoSeatGame(t, s.l, "Came back")
	s.at(36*time.Hour + time.Minute)
	if e := s.entry(t, meta.ID); e.idleWarnedAt.IsZero() {
		t.Fatal("not warned at 36 h")
	}
	s.at(40 * time.Hour)
	s.move(t, meta.ID, alice)
	s.at(48*time.Hour + time.Minute)
	if s.archived(t, meta.ID) {
		t.Fatal("closed after a move answered the warning")
	}
	if e := s.entry(t, meta.ID); !e.idleWarnedAt.IsZero() {
		t.Error("the move did not clear the warning")
	}
	s.at(40*time.Hour + 36*time.Hour + time.Minute) // warned again
	s.at(40*time.Hour + 48*time.Hour + time.Minute)
	if !s.archived(t, meta.ID) {
		t.Fatal("still open 48 hours after the last move")
	}
}

// ADR 0211 §6 (sub-PR 4): the warning reaches the table as
// idle_close_at on the view and the notifier as one TableIdleWarning;
// a person's move takes both back at once, not at the next sweep.
func TestIdleWarningStampsTheViewAndTellsTheNotifierOnce(t *testing.T) {
	s := newIdleStack(t, "", idleLimits())
	ev := &recordEvents{}
	s.l.SetTableEvents(ev)
	meta, alice, _ := startTwoSeatGame(t, s.l, "Warned")
	room := s.room(t, meta.ID)

	s.at(35 * time.Hour)
	if got := room.ViewFor(alice.String()).IdleCloseAt; got != "" {
		t.Fatalf("idle_close_at %q before the warning", got)
	}
	s.at(36*time.Hour + time.Minute)
	want := s.entry(t, meta.ID).idleCloseAt
	if want.IsZero() {
		t.Fatal("not warned at 36 h")
	}
	view, _, err := room.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if view.IdleCloseAt != want.UTC().Format(time.RFC3339) {
		t.Errorf("idle_close_at %q, want %v", view.IdleCloseAt, want)
	}
	if tbl, _ := s.l.NotifyTables().Table(meta.ID); !tbl.IdleCloseAt.Equal(want) {
		t.Errorf("the notifier reads the closing time as %v, want %v", tbl.IdleCloseAt, want)
	}
	s.at(37 * time.Hour)
	if w := ev.warnings(); len(w) != 1 || w[0].game != meta.ID || !w[0].closesAt.Equal(want) || w[0].hours != 36 {
		t.Fatalf("the notifier was told %+v, want one 36-hour warning closing at %v", w, want)
	}

	s.at(40 * time.Hour)
	s.move(t, meta.ID, alice)
	if got := room.ViewFor(alice.String()).IdleCloseAt; got != "" {
		t.Errorf("idle_close_at %q after a move, before the next sweep", got)
	}
	if tbl, _ := s.l.NotifyTables().Table(meta.ID); !tbl.IdleCloseAt.IsZero() {
		t.Error("the notifier still reads the warning after a move")
	}
	s.at(40*time.Hour + 10*time.Minute)
	if got := room.ViewFor(alice.String()).IdleCloseAt; got != "" {
		t.Errorf("the sweep put idle_close_at %q back after a move", got)
	}
	if len(ev.warnings()) != 1 {
		t.Errorf("warnings told: %d, want 1", len(ev.warnings()))
	}
}

// A lobby touch (a take-back, an away answer, a DM invite) answers the
// warning too, and its notice goes with it.
func TestIdleWarningIsTakenBackByATouch(t *testing.T) {
	s := newIdleStack(t, "", idleLimits())
	meta, _, _ := startTwoSeatGame(t, s.l, "Touched")
	room := s.room(t, meta.ID)
	s.at(36*time.Hour + time.Minute)
	if room.IdleCloseAt().IsZero() {
		t.Fatal("not warned at 36 h")
	}
	s.l.TouchTable(meta.ID)
	if !room.IdleCloseAt().IsZero() {
		t.Error("a touch left the table's notice up")
	}
	if e := s.entry(t, meta.ID); !e.idleWarnedAt.IsZero() {
		t.Error("a touch left the sweep's warning standing")
	}
}

func TestIdleGameCloseWaitsAnHourAfterALateWarning(t *testing.T) {
	s := newIdleStack(t, "", idleLimits())
	meta, _, _ := startTwoSeatGame(t, s.l, "Server was down")
	// The first sweep anyone sees is long past the deadline (the warning
	// was lost to a restart, say): warn now, close an hour later.
	s.at(60 * time.Hour)
	if s.archived(t, meta.ID) {
		t.Fatal("closed without a warning")
	}
	s.at(60*time.Hour + 30*time.Minute)
	if s.archived(t, meta.ID) {
		t.Fatal("closed within an hour of its warning")
	}
	s.at(61*time.Hour + time.Minute)
	if !s.archived(t, meta.ID) {
		t.Fatal("still open an hour after the warning")
	}
}

func TestIdleBotOnlyGameClosesAfterAnHourAlone(t *testing.T) {
	s := newIdleStack(t, "", idleLimits())
	meta, _ := s.l.Create("Bots only")
	for _, n := range []string{"Bot A", "Bot B"} {
		if _, _, err := s.l.AddBot(meta.ID, n, "random", "", "", botDeck(20)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.l.Start(meta.ID); err != nil {
		t.Fatal(err)
	}
	room := s.room(t, meta.ID)
	if room.PlayingOut() {
		t.Fatal("playing out before the sweep has seen the table")
	}
	s.at(time.Minute) // the sweep first sees it alone
	// Sub-PR 5 (owner decision 2): its bots play it out fast.
	if !room.PlayingOut() {
		t.Fatal("a game left to its bots is not played out fast")
	}
	s.at(time.Hour)
	if s.archived(t, meta.ID) {
		t.Fatal("closed before an hour alone")
	}
	// The fake bot host runs no bots, so this game has stalled: the
	// one-hour close is its backstop.
	s.at(time.Hour + 2*time.Minute)
	if !s.archived(t, meta.ID) {
		t.Fatal("a game left to its bots is still open after an hour")
	}
	if room.PlayingOut() {
		t.Error("a closed table is still playing out")
	}
	if got := s.entry(t, meta.ID).closedReason; got != ClosedIdleBots {
		t.Errorf("closed_reason %q, want %q", got, ClosedIdleBots)
	}
}

func TestIdleBotOnlyGameIsHeldByAWatchingSeat(t *testing.T) {
	s := newIdleStack(t, "", idleLimits())
	meta, _ := s.l.Create("Alice watches")
	_, alice, _ := s.l.Join(meta.ID, meta.InviteToken, "Alice")
	for _, n := range []string{"Bot A", "Bot B"} {
		if _, _, err := s.l.AddBot(meta.ID, n, "random", "", "", botDeck(20)); err != nil {
			t.Fatal(err)
		}
	}
	startWithDecks(t, s.l, meta, alice)
	if err := s.l.RoomOf(meta.ID).Game.Concede(alice); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	// Alice conceded but is still at the table (the fake reads her as
	// present), watching the bots finish.
	room := s.room(t, meta.ID)
	s.at(time.Minute)
	s.at(2 * time.Hour)
	if s.archived(t, meta.ID) {
		t.Fatal("closed while a conceded player watched")
	}
	if room.PlayingOut() {
		t.Fatal("the bots play out fast while a conceded player watches")
	}
	s.away.set(alice, protocol.PresenceOffline, s.start.Add(2*time.Hour))
	s.at(2*time.Hour + time.Minute)
	if !room.PlayingOut() {
		t.Fatal("not played out once the last watcher left")
	}
	// She looks in again: the table's own pace is back on the next sweep
	// (and at once on her socket, ws.Hub.ServeWS).
	s.away.set(alice, "", time.Time{})
	s.at(2*time.Hour + 6*time.Minute)
	if room.PlayingOut() {
		t.Fatal("still playing out with a person back at the table")
	}
	s.away.set(alice, protocol.PresenceOffline, s.start.Add(2*time.Hour+7*time.Minute))
	s.at(2*time.Hour + 8*time.Minute)
	if !room.PlayingOut() {
		t.Fatal("not played out again once she left")
	}
	s.at(3*time.Hour + 10*time.Minute)
	if !s.archived(t, meta.ID) {
		t.Fatal("still open an hour after the last watcher left")
	}
}

func TestIdleNothingClosesWithinAnHourOfResidency(t *testing.T) {
	lim := idleLimits()
	lim.IdleLobbyAfter = 10 * time.Minute
	s := newIdleStack(t, "", lim)
	meta, _ := s.l.Create("Fresh")
	s.at(30 * time.Minute)
	if s.archived(t, meta.ID) {
		t.Fatal("closed within an hour of becoming resident")
	}
	s.at(time.Hour + time.Minute)
	if !s.archived(t, meta.ID) {
		t.Fatal("not closed once the grace ended")
	}
}

func TestIdleZeroTurnsEachCleanupOff(t *testing.T) {
	s := newIdleStack(t, "", capacity.Limits{})
	lobbyTable, _ := s.l.Create("Lobby")
	game1, _, _ := startTwoSeatGame(t, s.l, "Game")
	s.at(1000 * time.Hour)
	if s.archived(t, lobbyTable.ID) || s.archived(t, game1.ID) {
		t.Fatal("a cleanup that is off closed a table")
	}
}

func TestIdleUnarchiveGivesAFreshClock(t *testing.T) {
	s := newIdleStack(t, "", idleLimits())
	meta, _ := s.l.Create("Asked back")
	s.at(24*time.Hour + time.Minute)
	if !s.archived(t, meta.ID) {
		t.Fatal("not closed")
	}
	if _, err := s.l.SetArchived(meta.ID, false); err != nil {
		t.Fatal(err)
	}
	if got := s.entry(t, meta.ID).closedReason; got != "" {
		t.Errorf("unarchive left closed_reason %q", got)
	}
	s.at(24*time.Hour + 10*time.Minute)
	if s.archived(t, meta.ID) {
		t.Fatal("the sweep closed the table again straight after an unarchive")
	}
}

func TestIdleClosedTableIsUnloadedAfterTheKeepWindow(t *testing.T) {
	dir := t.TempDir()
	s := newIdleStack(t, dir, idleLimits())
	meta, _, _ := startTwoSeatGame(t, s.l, "Gone for good")
	s.at(36*time.Hour + time.Minute)
	s.at(48*time.Hour + time.Minute)
	if !s.archived(t, meta.ID) {
		t.Fatal("not closed")
	}
	restorePoint := filepath.Join(dir, "restore", meta.ID.String()+".json")
	replay := filepath.Join(dir, "replays", meta.ID.String()+".jsonl")
	if _, err := os.Stat(replay); err != nil {
		t.Fatalf("no replay to keep: %v", err)
	}

	s.at(48*time.Hour + IdleClosedKeep)
	if _, err := s.l.Get(meta.ID); err != nil {
		t.Fatalf("unloaded inside the keep window: %v", err)
	}
	s.at(48*time.Hour + IdleClosedKeep + 2*time.Minute)
	if _, err := s.l.Get(meta.ID); err != ErrGameNotFound {
		t.Fatalf("still resident after the keep window: %v", err)
	}
	if s.mgr.Get(meta.ID) != nil {
		t.Error("the room is still in the manager")
	}
	if _, err := os.Stat(restorePoint); !os.IsNotExist(err) {
		t.Errorf("the restore point is still on disk: %v", err)
	}
	if _, err := os.Stat(replay); err != nil {
		t.Errorf("the replay went with the room: %v", err)
	}
	rec, _, err := s.l.store.LoadGame(context.Background(), meta.ID)
	if err != nil {
		t.Fatal(err)
	}
	if rec.State != string(game.StateEnded) || rec.EndedAt == nil || rec.Outcome != "" || rec.ClosedReason != ClosedIdleGame || rec.ArchivedAt == nil {
		t.Errorf("row after the unload: %+v; want ended, archived, idle_game, no outcome", rec)
	}
}

// A row the sweep has written comes back with its last activity, and a
// row from before migration 0031 starts its clock at the restore.
func TestIdleLastActivitySurvivesARestart(t *testing.T) {
	dir := t.TempDir()
	s := newIdleStack(t, dir, idleLimits())
	meta, _ := s.l.Create("Restarted")
	if _, _, err := s.l.Join(meta.ID, meta.InviteToken, "Alice"); err != nil {
		t.Fatal(err)
	}
	s.at(10 * time.Hour)
	rec, _, err := s.l.store.LoadGame(context.Background(), meta.ID)
	if err != nil {
		t.Fatal(err)
	}
	if rec.LastActiveAt == nil || rec.LastActiveAt.Before(s.start) || rec.LastActiveAt.After(s.start.Add(time.Second)) {
		t.Fatalf("last_active_at %v, want the join at %v", rec.LastActiveAt, s.start)
	}

	// --- the deploy -----------------------------------------------
	s2 := newIdleStack(t, dir, idleLimits())
	s2.mu.Lock()
	s2.start, s2.now = s.start, s.start.Add(10*time.Hour)
	s2.mu.Unlock()
	if n := s2.l.RestoreFromDisk(quietLogger()); n != 1 {
		t.Fatalf("restored %d games, want 1", n)
	}
	e := s2.entry(t, meta.ID)
	if e.lastActiveStored == nil || !e.lastActiveStored.Equal(*rec.LastActiveAt) {
		t.Fatalf("restored last activity %v, want %v", e.lastActiveStored, rec.LastActiveAt)
	}
	if !e.lastTouched.IsZero() {
		t.Errorf("a row with a last activity was touched at load: %v", e.lastTouched)
	}
	s2.at(24*time.Hour + time.Minute)
	if !s2.archived(t, meta.ID) {
		t.Fatal("the restored clock did not carry the stored last activity")
	}
}
