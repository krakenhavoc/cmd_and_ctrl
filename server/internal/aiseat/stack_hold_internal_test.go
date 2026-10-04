package aiseat

import (
	"context"
	"io"
	"log/slog"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// stack_hold_internal_test.go pins ADR 0119 §2's bot side: the bot
// speed presets carry a stack hold, and holdForStack keeps a pass
// waiting until another seat's top item has been on the stack for it.
// In-package because the hold is unexported, and synchronous: every
// test here runs holdForStack on its own goroutine, and the one that
// changes the stack mid-hold does it through the runner's poll hook,
// so nothing here waits on a goroutine it does not schedule.

// holdBackstop bounds a hold that should have returned at once. It is
// a backstop, not an assertion: each test asserts the context is still
// live afterwards, which says the hold returned without running it out.
const holdBackstop = 30 * time.Second

// holdTable is a two-seat table past the mulligans, at a step where
// the active seat holds priority. It returns the room, the active seat
// and the other one.
func holdTable(t *testing.T) (*ws.Room, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < 2; i++ {
		deck := make([]game.Card, 0, 20)
		for j := 0; j < 20; j++ {
			c := game.NewCard("Mountain", uuid.Nil)
			c.TypeLine = "Basic Land — Mountain"
			deck = append(deck, c)
		}
		if _, err := g.AddPlayer("Seat", deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(119, 2))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	for i := 0; g.Turn.PriorityHolder == game.NoPriority; i++ {
		if i >= 40 {
			t.Fatal("no step granted priority within 40 advances")
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	active := g.Seats[g.Turn.ActiveSeat].ID
	other := g.Seats[1-g.Turn.ActiveSeat].ID
	return ws.NewRoom(g, slog.New(slog.NewTextHandler(io.Discard, nil)), ""), active, other
}

// announce puts a triggered ability controlled by `who` on top of the
// stack and returns its ID.
func announce(t *testing.T, g *game.Game, who uuid.UUID) uuid.UUID {
	t.Helper()
	p := g.PlayerByID(who)
	if p == nil || p.Hand == nil || len(p.Hand.Cards) == 0 {
		t.Fatal("the announcing seat has no hand card to be the source")
	}
	if err := g.AnnounceTrigger(who, p.Hand.Cards[0].InstanceID, game.AbilityParams{Label: "a test trigger"}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	sp := g.StackPresenceSnapshot()
	if sp.TopController != who {
		t.Fatalf("the announced trigger is not on top of the stack for its controller: %+v", sp)
	}
	return sp.Top
}

func TestBotPacePresetsCarryAStackHold(t *testing.T) {
	want := map[game.BotPace]time.Duration{
		game.BotPaceFast:   0,
		game.BotPaceNormal: 2 * time.Second,
		game.BotPaceSlow:   3 * time.Second,
	}
	for pace, hold := range want {
		if got := botPacePresets[pace].StackHold; got != hold {
			t.Errorf("%s: StackHold = %v, want %v (ADR 0119 §2)", pace, got, hold)
		}
	}
}

func TestStackHoldNowFollowsTheTablePace(t *testing.T) {
	room := paceRoom(t)
	r := &Runner{room: room, cfg: Config{FollowTablePace: true}}
	for _, pace := range []game.BotPace{game.BotPaceFast, game.BotPaceNormal, game.BotPaceSlow} {
		p := pace
		if err := room.Game.UpdateSettings(uuid.Nil, game.SettingsPatch{BotPace: &p}); err != nil {
			t.Fatalf("UpdateSettings: %v", err)
		}
		if got, want := r.stackHoldNow(), botPacePresets[pace].StackHold; got != want {
			t.Errorf("%s: stackHoldNow() = %v, want %v", pace, got, want)
		}
	}
}

// A hand-built Config (FollowTablePace off) and a stepped runner never
// hold, so tests, arenas and the soak keep their speed.
func TestStackHoldIsOffWithoutTablePaceAndWhenStepped(t *testing.T) {
	room := paceRoom(t)
	slow := game.BotPaceSlow
	if err := room.Game.UpdateSettings(uuid.Nil, game.SettingsPatch{BotPace: &slow}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	hand := &Runner{room: room, cfg: Config{}}
	if got := hand.stackHoldNow(); got != 0 {
		t.Errorf("FollowTablePace off: stackHoldNow() = %v, want 0", got)
	}
	stepped := NewStepped(room, uuid.New(), NewRandomPolicy(rand.NewPCG(1, 2)), DefaultConfig(), nil)
	if got := stepped.stackHoldNow(); got != 0 {
		t.Errorf("stepped runner on a slow table: stackHoldNow() = %v, want 0", got)
	}
}

func TestNoteStackKeepsFirstSeenAndForgetsWhatLeft(t *testing.T) {
	r := &Runner{}
	a, b := uuid.New(), uuid.New()
	t0 := time.Unix(1000, 0)
	r.noteStackPresence(game.StackPresence{Items: []uuid.UUID{a}}, t0)
	r.noteStackPresence(game.StackPresence{Items: []uuid.UUID{a, b}}, t0.Add(time.Second))
	if got := r.stackSeen[a]; !got.Equal(t0) {
		t.Errorf("first-seen for an item still on the stack moved to %v, want %v", got, t0)
	}
	if got := r.stackSeen[b]; !got.Equal(t0.Add(time.Second)) {
		t.Errorf("a new item's first-seen = %v, want the frame it appeared in", got)
	}
	r.noteStackPresence(game.StackPresence{Items: []uuid.UUID{b}}, t0.Add(2*time.Second))
	if _, ok := r.stackSeen[a]; ok {
		t.Error("an item that left the stack is still remembered")
	}

	room := paceRoom(t)
	r = &Runner{room: room, stackSeen: map[uuid.UUID]time.Time{a: t0}}
	r.noteStack(0, t0)
	if r.stackSeen != nil {
		t.Error("with no hold the runner still keeps first-seen times")
	}
}

func TestHoldForStackWaitsOutAnotherSeatsItem(t *testing.T) {
	room, active, other := holdTable(t)
	announce(t, room.Game, active)
	r := &Runner{room: room, seat: other}

	const hold = 150 * time.Millisecond
	started := time.Now()
	if !r.holdForStack(context.Background(), hold) {
		t.Fatal("holdForStack reported a changed stack; nothing changed it")
	}
	// First-seen is stamped inside the call, so the hold cannot end
	// before started+hold. A lower bound only: a loaded machine makes
	// it longer, never shorter.
	if elapsed := time.Since(started); elapsed < hold {
		t.Errorf("the pass went after %v, before the %v hold was over", elapsed, hold)
	}
}

func TestHoldForStackMeasuresFromFirstSeen(t *testing.T) {
	room, active, other := holdTable(t)
	top := announce(t, room.Game, active)
	// The runner saw the item two hours ago, so an hour's hold is over.
	r := &Runner{room: room, seat: other, stackSeen: map[uuid.UUID]time.Time{top: time.Now().Add(-2 * time.Hour)}}
	ctx, cancel := context.WithTimeout(context.Background(), holdBackstop)
	defer cancel()
	if !r.holdForStack(ctx, time.Hour) {
		t.Fatal("holdForStack reported a changed stack; nothing changed it")
	}
	if ctx.Err() != nil {
		t.Fatal("the hold ran from the call, not from when the runner first saw the item")
	}
}

func TestHoldForStackNeverHoldsTheSeatsOwnItem(t *testing.T) {
	room, active, _ := holdTable(t)
	announce(t, room.Game, active)
	r := &Runner{room: room, seat: active}
	ctx, cancel := context.WithTimeout(context.Background(), holdBackstop)
	defer cancel()
	if !r.holdForStack(ctx, time.Hour) {
		t.Fatal("holdForStack reported a changed stack; nothing changed it")
	}
	if ctx.Err() != nil {
		t.Fatal("the runner held a pass on its own item (CR 117.3c: it passes at once)")
	}
}

func TestHoldForStackNeverHoldsAnEmptyStack(t *testing.T) {
	room, _, other := holdTable(t)
	r := &Runner{room: room, seat: other}
	ctx, cancel := context.WithTimeout(context.Background(), holdBackstop)
	defer cancel()
	if !r.holdForStack(ctx, time.Hour) || ctx.Err() != nil {
		t.Fatal("the runner held a pass on an empty stack")
	}
}

// If the top item changes while the runner holds, the pass it decided
// on the old frame must not go: holdForStack says so and the runner
// re-enumerates.
func TestHoldForStackReportsATopItemThatChanged(t *testing.T) {
	room, active, other := holdTable(t)
	first := announce(t, room.Game, active)
	r := &Runner{room: room, seat: other}
	polls := 0
	r.stackHoldPolled = func() {
		polls++
		if polls == 1 {
			if second := announce(t, room.Game, active); second == first {
				t.Fatal("the second trigger did not go on top of the first")
			}
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), holdBackstop)
	defer cancel()
	if r.holdForStack(ctx, time.Hour) {
		t.Fatal("holdForStack let the pass go after the top of the stack changed under it")
	}
	if ctx.Err() != nil {
		t.Fatal("holdForStack ran out its hold instead of noticing the new top item")
	}
}
