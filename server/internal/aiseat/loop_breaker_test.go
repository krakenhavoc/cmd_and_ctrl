package aiseat_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// loop_breaker_test.go is the bot half of #628 (CR 726).
//
// A bot's pass is as automatic as a browser's autopass toggle, so a
// table of bots spinning on a real trigger loop is the same runaway
// the breaker exists to stop — and a bot-only table has no human to
// notice. This test builds the loop for real and asserts that the
// bots stop feeding it once the engine raises the notice.
//
// It is NOT behind AISEAT_GAME_TESTS: it plays no whole game, it
// runs in well under a second, and it is the only coverage of the
// runner's one line of loop-breaker code.

const (
	loopOracleA = "test-bot-loop-a"
	loopOracleB = "test-bot-loop-b"
	loopTokenA  = "Bot Loop Spark"
	loopTokenB  = "Bot Loop Shard"
)

// installTriggerLoop points CatalogTriggers at two synthetic
// permanents that make each other's tokens, restoring the catalog's
// own hook when the test ends. Unknown oracle IDs fall through to the
// real catalog so the bots' Mountains and Lightning Bolts still work.
func installTriggerLoop(t *testing.T) {
	t.Helper()
	prev := game.CatalogTriggers
	t.Cleanup(func() { game.CatalogTriggers = prev })
	game.CatalogTriggers = func(id string) []game.TriggeredAbility {
		switch id {
		case loopOracleA:
			return []game.TriggeredAbility{botLoopTrigger(loopTokenB, "Loop Engine A — create a Spark", loopTokenA)}
		case loopOracleB:
			return []game.TriggeredAbility{botLoopTrigger(loopTokenA, "Loop Engine B — create a Shard", loopTokenB)}
		}
		if prev != nil {
			return prev(id)
		}
		return nil
	}
}

func botLoopTrigger(watch, label, make string) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventETB},
		AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
			for _, c := range g.Battlefield.Cards {
				if c.InstanceID == ev.CardID {
					return c.Name == watch
				}
			}
			return false
		},
		Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			controller, owner := source.Controller, source.Owner
			return game.NewTriggeredItem(source, label, func(g *game.Game, _ *game.StackItem) error {
				pushLoopToken(g, make, controller, owner)
				return nil
			})
		},
	}
}

// pushLoopToken puts one synthetic token on the battlefield and
// announces it. Caller must hold the game's write lock.
func pushLoopToken(g *game.Game, name string, controller, owner uuid.UUID) {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Token Creature — Elemental",
		Power:      1,
		Toughness:  1,
		Owner:      owner,
		Controller: controller,
	})
	g.EmitEvent(game.Event{Kind: game.EventETB, CardID: id, Actor: controller})
}

// TestBotsStopPassingWhenTheLoopBreakerFires: two bot seats, a real
// two-permanent trigger loop, and a threshold of five. The bots pass
// the loop around until the engine raises the notice, and then they
// hold — the room's commit sequence stops moving, which is the whole
// point. The alternative was a table that spins until the process is
// killed.
func TestBotsStopPassingWhenTheLoopBreakerFires(t *testing.T) {
	installTriggerLoop(t)
	room := newRoom(t, 2, 21)
	g := room.Game
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g.WithWriteLock(func() { g.LoopThreshold = 5 })

	pol := func() aiseat.Policy { return &scripted{prefer: []string{"Keep hand"}} }
	aiseat.Start(ctx, room, g.Seats[0].ID, pol(), aiseat.Config{}, nil, testLogger())
	aiseat.Start(ctx, room, g.Seats[1].ID, pol(), aiseat.Config{}, nil, testLogger())
	waitFor(t, "both seats to keep", 3*time.Second, func() bool {
		return handKept(g, g.Seats[0].ID) && handKept(g, g.Seats[1].ID)
	})

	// Start the loop turning. The first token's ETB triggers engine
	// B; from there every resolution queues the next trigger, so the
	// stack is never empty again and the turn can never advance —
	// which is exactly the shape the breaker exists for.
	owner := g.Seats[0].ID
	g.WithWriteLock(func() {
		for _, c := range []game.Card{
			{InstanceID: uuid.New(), Name: "Loop Engine A", OracleID: loopOracleA,
				TypeLine: "Artifact", Owner: owner, Controller: owner},
			{InstanceID: uuid.New(), Name: "Loop Engine B", OracleID: loopOracleB,
				TypeLine: "Artifact", Owner: owner, Controller: owner},
		} {
			g.Battlefield.PushTop(c)
		}
		pushLoopToken(g, loopTokenA, owner, owner)
	})

	waitFor(t, "the loop breaker to fire", 5*time.Second, func() bool {
		return g.AutoPassSuspended()
	})
	notice := g.CurrentLoopNotice()
	if notice == nil || notice.Count < 5 {
		t.Fatalf("loop notice = %+v, want a count of at least the threshold", notice)
	}

	// With automatic passing suspended, nothing commits. A quarter of
	// a second is many wakes at the runner's pacing; before this
	// change the same window carried dozens of passes.
	settled := room.Seq()
	time.Sleep(250 * time.Millisecond)
	if got := room.Seq(); got != settled {
		t.Errorf("room sequence moved from %d to %d while the loop breaker was up — a bot kept passing", settled, got)
	}
	if !g.AutoPassSuspended() {
		t.Error("the notice cleared itself; only a player decision or a new turn should")
	}
}
