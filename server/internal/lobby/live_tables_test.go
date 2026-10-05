package lobby

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// Lobby.LiveTables, the admin views' copy of the tables in memory (ADR
// 0124 §5). TestScrapeDuringTableTraffic runs it beside the metrics
// accessors under -race; these pin what it copies and the locks it
// does not take.

// writeCountingStore counts every write the lobby makes to its store.
type writeCountingStore struct {
	Store
	writes atomic.Int64
}

func (s *writeCountingStore) CreateGame(ctx context.Context, g GameRecord, inv []InviteRecord) error {
	s.writes.Add(1)
	return s.Store.CreateGame(ctx, g, inv)
}

func (s *writeCountingStore) UpdateGame(ctx context.Context, g GameRecord) error {
	s.writes.Add(1)
	return s.Store.UpdateGame(ctx, g)
}

func (s *writeCountingStore) ReplaceSeats(ctx context.Context, id uuid.UUID, seats []SeatRecord) error {
	s.writes.Add(1)
	return s.Store.ReplaceSeats(ctx, id, seats)
}

func (s *writeCountingStore) DeleteGame(ctx context.Context, id uuid.UUID) error {
	s.writes.Add(1)
	return s.Store.DeleteGame(ctx, id)
}

func liveTableByID(t *testing.T, tables []LiveTable, id uuid.UUID) LiveTable {
	t.Helper()
	for _, lt := range tables {
		if lt.ID == id {
			return lt
		}
	}
	t.Fatalf("LiveTables has no table %s", id)
	return LiveTable{}
}

// Every kind of seat and table, copied with its fields.
func TestLiveTablesCopiesEveryTableAndSeat(t *testing.T) {
	l, _, _ := newPracticeLobby(t)

	waiting, err := l.Create("Waiting")
	if err != nil {
		t.Fatal(err)
	}
	_, alice, err := l.Join(waiting.ID, waiting.InviteToken, "Alice")
	if err != nil {
		t.Fatal(err)
	}
	bobUser := uuid.New()
	_, bob, err := l.JoinAs(waiting.ID, waiting.InviteToken, "Bob",
		DiscordIdentity{ID: "d-bob", Username: "bob", GlobalName: "Bobby"}, bobUser)
	if err != nil {
		t.Fatal(err)
	}
	_, agent, err := l.JoinAgent(waiting.ID, waiting.InviteToken, "Claude", AgentDecl{Client: "Claude Code"})
	if err != nil {
		t.Fatal(err)
	}
	_, bot, err := l.AddBot(waiting.ID, "Bot 1", "random", "", "Mono Red", botDeck(20))
	if err != nil {
		t.Fatal(err)
	}

	running, _, _ := startTwoSeatGame(t, l, "Running")
	if _, err := l.SetArchived(running.ID, true); err != nil {
		t.Fatal(err)
	}
	human, pbot := practiceSeats("user:p")
	practice, _, err := l.CreatePractice(human, pbot)
	if err != nil {
		t.Fatal(err)
	}

	tables := l.LiveTables()
	if len(tables) != 3 {
		t.Fatalf("LiveTables has %d tables, want 3", len(tables))
	}

	w := liveTableByID(t, tables, waiting.ID)
	if w.Name != "Waiting" || w.State != string(game.StateLobby) || w.Archived || w.Practice {
		t.Errorf("waiting table = %+v", w)
	}
	if !w.CreatedAt.Equal(waiting.CreatedAt) || w.CreatedAt.IsZero() {
		t.Errorf("waiting table created %v, want %v", w.CreatedAt, waiting.CreatedAt)
	}
	want := map[uuid.UUID]LiveSeat{
		alice: {Seat: 0, PlayerID: alice, Kind: metrics.SeatHuman, Name: "Alice", Host: true},
		bob:   {Seat: 1, PlayerID: bob, Kind: metrics.SeatHuman, UserID: bobUser.String(), Name: "Bob", DisplayName: "Bobby"},
		agent: {Seat: 2, PlayerID: agent, Kind: metrics.SeatAgent, Name: "Claude", AgentClient: "claude-code"},
		bot:   {Seat: 3, PlayerID: bot, Kind: metrics.SeatBot, Name: "Bot 1", BotTier: "random"},
	}
	if len(w.Seats) != len(want) {
		t.Fatalf("waiting table has %d seats, want %d", len(w.Seats), len(want))
	}
	for _, s := range w.Seats {
		if s != want[s.PlayerID] {
			t.Errorf("seat %s = %+v, want %+v", s.Name, s, want[s.PlayerID])
		}
	}

	r := liveTableByID(t, tables, running.ID)
	if r.State != string(game.StateActive) || !r.Archived || r.Practice || len(r.Seats) != 2 {
		t.Errorf("running archived table = %+v", r)
	}
	p := liveTableByID(t, tables, practice.ID)
	if !p.Practice || p.State != string(game.StateActive) || len(p.Seats) != 2 {
		t.Errorf("practice table = %+v", p)
	}

	// The copy is the caller's: changing it changes nothing in the
	// lobby, and the lobby's later changes do not reach it.
	w.Seats[0].Name = "Mallory"
	if _, err := l.RemoveBot(waiting.ID, bot); err != nil {
		t.Fatal(err)
	}
	again := liveTableByID(t, l.LiveTables(), waiting.ID)
	if again.Seats[0].Name != "Alice" || len(again.Seats) != 3 {
		t.Errorf("after editing the copy and removing the bot: seat 0 %q, %d seats; want Alice and 3", again.Seats[0].Name, len(again.Seats))
	}
	if len(w.Seats) != 4 {
		t.Errorf("an earlier copy changed to %d seats", len(w.Seats))
	}
}

// LiveTables takes no room lock and releases l.mu: it answers while a
// commit holds the room, and the lobby is free when it returns. It
// writes nothing to the store.
func TestLiveTablesTakesNoRoomLockAndWritesNothing(t *testing.T) {
	store := &writeCountingStore{Store: NewMemoryStore()}
	l := NewLobbyWithStore(ws.NewRoomManager(quietLogger(), ""), store)
	meta, alice, _ := startTwoSeatGame(t, l, "Locks")
	room := l.RoomOf(meta.ID)

	// Inside a commit: the room's mutex is held until fn returns.
	errRollback := errors.New("roll back")
	_, _, err := room.Apply(alice, func() error {
		done := make(chan []LiveTable, 1)
		go func() { done <- l.LiveTables() }()
		select {
		case got := <-done:
			if len(got) != 1 {
				t.Errorf("LiveTables inside a commit = %d tables, want 1", len(got))
			}
		case <-time.After(5 * time.Second):
			t.Error("LiveTables blocked on the room's lock")
		}
		return errRollback
	})
	if !errors.Is(err, errRollback) {
		t.Fatalf("Apply: %v", err)
	}
	if !l.mu.TryLock() {
		t.Fatal("LiveTables left l.mu held")
	}
	l.mu.Unlock()

	// Nothing else is changing the table, so any write now would be
	// LiveTables' own.
	before := store.writes.Load()
	for range 10 {
		_ = l.LiveTables()
	}
	if n := store.writes.Load() - before; n != 0 {
		t.Errorf("LiveTables made %d store writes, want none", n)
	}
}
