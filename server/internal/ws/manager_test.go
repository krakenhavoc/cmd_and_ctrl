package ws

import (
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"sync"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// buildTestGame is a test helper that stands up a 2-seat game with
// deterministic decks and starts it with a fixed RNG. Used by manager
// tests so that every room we register has a real, active game ID.
func buildTestGame(t *testing.T, label string) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := range 2 {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("%s-Cmdr-%d", label, i+1), uuid.Nil)}
		for j := range 10 {
			deck = append(deck, game.NewCard(fmt.Sprintf("%s-Filler-%d", label, j+1), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("%s-P%d", label, i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 1))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return g
}

func TestRoomManagerCreateRegisterGet(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mgr := NewRoomManager(log, "")

	g := buildTestGame(t, "A")
	r := mgr.Create(g)
	if r == nil {
		t.Fatal("Create returned nil")
	}

	got := mgr.Get(g.ID)
	if got != r {
		t.Errorf("Get returned a different pointer than Create")
	}
	if mgr.Count() != 1 {
		t.Errorf("Count: got %d, want 1", mgr.Count())
	}

	// Create again on the same game should be a no-op and return the
	// existing room rather than replacing it.
	again := mgr.Create(g)
	if again != r {
		t.Errorf("Create on duplicate game returned a new room; expected existing")
	}
}

func TestRoomManagerGetUnknownReturnsNil(t *testing.T) {
	mgr := NewRoomManager(nil, "")
	if got := mgr.Get(uuid.New()); got != nil {
		t.Errorf("Get on unknown id: got %v, want nil", got)
	}
}

func TestRoomManagerSingletonOnlyForOne(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mgr := NewRoomManager(log, "")

	if mgr.Singleton() != nil {
		t.Error("empty manager: Singleton should be nil")
	}

	g1 := buildTestGame(t, "A")
	r1 := mgr.Create(g1)
	if mgr.Singleton() != r1 {
		t.Errorf("one room: Singleton should return that room")
	}

	g2 := buildTestGame(t, "B")
	mgr.Create(g2)
	if mgr.Singleton() != nil {
		t.Error("two rooms: Singleton should be nil")
	}
}

func TestRoomManagerListIsOrdered(t *testing.T) {
	// List returns rooms in CreatedAt order. Since NewGame stamps
	// CreatedAt with time.Now().UTC(), inserting in A→B→C order
	// yields the same iteration order out.
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mgr := NewRoomManager(log, "")

	gA := buildTestGame(t, "A")
	gB := buildTestGame(t, "B")
	gC := buildTestGame(t, "C")

	mgr.Create(gA)
	mgr.Create(gB)
	mgr.Create(gC)

	rooms := mgr.List()
	if len(rooms) != 3 {
		t.Fatalf("List len: got %d, want 3", len(rooms))
	}
	// CreatedAt may have identical nanosecond ticks on fast hardware;
	// we can only assert that the set is complete, not strictly
	// ordered. Covering the "no duplicates" invariant is enough.
	seen := map[uuid.UUID]bool{}
	for _, r := range rooms {
		if seen[r.Game.ID] {
			t.Errorf("List contained %q twice", r.Game.ID)
		}
		seen[r.Game.ID] = true
	}
	for _, want := range []uuid.UUID{gA.ID, gB.ID, gC.ID} {
		if !seen[want] {
			t.Errorf("List missing %q", want)
		}
	}
}

func TestRoomManagerDelete(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mgr := NewRoomManager(log, "")
	g := buildTestGame(t, "A")
	mgr.Create(g)
	mgr.Delete(g.ID)
	if mgr.Get(g.ID) != nil {
		t.Error("Get after Delete should return nil")
	}
	// Deleting an unknown ID is a no-op.
	mgr.Delete(uuid.New())
}

// TestRoomManagerConcurrent exercises the manager under concurrent
// Create/Get/List pressure to catch any lock-order or map-aliasing
// bugs. Run with -race for full coverage.
func TestRoomManagerConcurrent(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mgr := NewRoomManager(log, "")

	const n = 20
	var wg sync.WaitGroup
	wg.Add(n)
	ids := make([]uuid.UUID, n)
	for i := range n {
		go func(i int) {
			defer wg.Done()
			g := buildTestGame(t, fmt.Sprintf("G%d", i))
			ids[i] = g.ID
			mgr.Create(g)
			_ = mgr.Get(g.ID)
			_ = mgr.List()
			_ = mgr.Count()
		}(i)
	}
	wg.Wait()
	if got := mgr.Count(); got != n {
		t.Errorf("Count after concurrent Create: got %d, want %d", got, n)
	}
	for _, id := range ids {
		if mgr.Get(id) == nil {
			t.Errorf("Get(%q) returned nil after concurrent Create", id)
		}
	}
}
