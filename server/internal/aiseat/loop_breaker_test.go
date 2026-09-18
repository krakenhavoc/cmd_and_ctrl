package aiseat_test

import (
	"context"
	"testing"

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
// the loop around until the engine raises the notice, take the CR 726
// shortcut once, and then stop — the room's commit sequence stops
// moving, which is the whole point. The alternative was a table that
// spins until the process is killed.
//
// "They stopped" used to be a 250ms sleep and a sequence that had not
// moved (#848). It is now the runners' own answer: a bot that kept
// passing would never PARK, because every pass commits and every
// commit wakes it, so both seats reaching Idle is the assertion — and
// once they are parked with nothing else writing to this room, the
// sequence has stopped by construction rather than by hoping.
func TestBotsStopPassingWhenTheLoopBreakerFires(t *testing.T) {
	installTriggerLoop(t)
	room := newRoom(t, 2, 21)
	g := room.Game
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const threshold = 5
	g.WithWriteLock(func() { g.LoopThreshold = threshold })

	pol := func() aiseat.Policy { return &scripted{prefer: []string{"Keep hand"}} }
	r0 := aiseat.Start(ctx, room, g.Seats[0].ID, pol(), aiseat.Config{}, nil, testLogger())
	r1 := aiseat.Start(ctx, room, g.Seats[1].ID, pol(), aiseat.Config{}, nil, testLogger())
	waitFor(t, "both seats to keep", func() bool {
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

	// #804 changed what "they stop" looks like, and the change is the
	// point of the feature. The breaker now comes with a CR 726
	// prompt, so a bot seat that controls the loop answers it rather
	// than sitting on a question: 10 more iterations the first time it
	// is asked in a turn, and stop the second time, because the
	// enumerator offers nothing but stop on a repeat ask
	// (internal/legal, docs/bot.md). The table therefore runs the
	// number it agreed to and then comes to rest — bounded, which is
	// all a bot-only table on a real loop can be.
	waitFor(t, "both bots to park with the notice up", func() bool {
		return r0.Idle() && r1.Idle() && g.AutoPassSuspended()
	})
	notice := g.CurrentLoopNotice()
	if notice == nil || notice.Count < threshold {
		t.Fatalf("loop notice = %+v, want a count of at least the threshold", notice)
	}
	// Both asks answered: nothing is left owed, which is why the
	// runners could park at all.
	var owed int
	g.ReadSnapshot(func() {
		for _, c := range g.PendingChoices {
			if c != nil {
				owed++
			}
		}
	})
	if owed != 0 {
		t.Errorf("%d prompts still open with both bots parked — a seat is sitting on a question", owed)
	}

	// The count is the assertion that the shortcut ran once and only
	// once: the threshold's worth of iterations that raised the first
	// notice, plus the K the bot agreed to, and then nothing.
	wantResolutions := threshold + game.DefaultLoopShortcutIterations
	var got int
	g.ReadSnapshot(func() {
		for _, n := range g.TurnTally.Resolved {
			if n > got {
				got = n
			}
		}
	})
	if got != wantResolutions {
		t.Errorf("the loop's busiest ability resolved %d times, want %d (threshold %d + one shortcut of %d)",
			got, wantResolutions, threshold, game.DefaultLoopShortcutIterations)
	}

	// Parked, with nothing owed and nothing else writing to this room,
	// the sequence has stopped by construction. Re-reading both is the
	// check that it has: a bot still passing would have woken itself.
	stopped := room.Seq()
	if !r0.Idle() || !r1.Idle() || room.Seq() != stopped {
		t.Errorf("room sequence moved past %d with the loop breaker up — a bot kept passing", stopped)
	}
	if !g.AutoPassSuspended() {
		t.Error("the notice cleared itself; only a player decision or a new turn should")
	}
}
