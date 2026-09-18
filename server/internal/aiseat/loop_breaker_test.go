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
//
// An ARTIFACT token, not a creature, and that is load-bearing (#924).
// The trigger watches EventETB and matches on the name, so nothing
// here needs a creature — but a creature token is a legal attacker,
// and this fixture pushes a card straight onto the battlefield rather
// than through the entry pipeline, so it never picks up the
// summoning-sickness flag that would stop it being one. A table that
// reached declare_attackers therefore handed the loop's controller a
// brand-new attacker every single resolution, and declaring an
// attacker is a player decision: notePlayerDecisionLocked restarted
// the run and cleared the notice, so the breaker never tripped again,
// the bots never parked, and the wait below burned its whole 30s
// budget. That was the second failure mode measured on develop — a
// fixture that gave the bots something to do, in a test whose premise
// is that they have nothing to do but pass.
func pushLoopToken(g *game.Game, name string, controller, owner uuid.UUID) {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Token Artifact",
		Owner:      owner,
		Controller: controller,
	})
	g.EmitEvent(game.Event{Kind: game.EventETB, CardID: id, Actor: controller})
}

// owedChoicesLocked counts the prompts nobody has answered yet — the
// engine's own "is this table waiting on somebody". Caller must hold
// the game's read lock.
func owedChoicesLocked(g *game.Game) int {
	n := 0
	for _, c := range g.PendingChoices {
		if c != nil {
			n++
		}
	}
	return n
}

// busiestResolutionsLocked is the resolution count of the ability that
// has resolved most this turn — the loop's own, in this fixture, since
// nothing else is resolving. Caller must hold the game's read lock.
func busiestResolutionsLocked(g *game.Game) int {
	n := 0
	for _, r := range g.TurnTally.Resolved {
		if r > n {
			n = r
		}
	}
	return n
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
// commit wakes it, so both seats reaching Idle is part of the
// assertion — and once they are parked with nothing else writing to
// this room, the sequence has stopped by construction rather than by
// hoping. The rest of it is the table's own state, which is #924: a
// parked runner is not on its own the same fact as a settled table.
func TestBotsStopPassingWhenTheLoopBreakerFires(t *testing.T) {
	installTriggerLoop(t)
	room := newRoom(t, 2, 21)
	g := room.Game
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const threshold = 5

	// The whole loop goes in BEFORE the runners exist, and that
	// ordering is the second half of #924.
	//
	// This used to seat the bots, wait for both hands to be kept, and
	// only then push the engines — three statements with two live bot
	// goroutines running between them. A bot seat is a goroutine the
	// test does not schedule, and these two need no I/O and think in
	// nanoseconds: on a single-CPU machine the test goroutine can lose
	// the processor the instant the keep-wait returns and not get it
	// back until the pair have played all 24 turns of a 2-seat game to
	// a winner, a thousand commits later. The engines then landed on a
	// finished table, nothing ever resolved, and the wait below sat
	// there for its full budget. Building the board first makes the
	// injection point a fact rather than a race: there is no other
	// goroutine yet.
	//
	// The first token's ETB triggers engine B; from there every
	// resolution queues the next trigger, so the table has nothing to
	// do but feed the loop — which is exactly the shape the breaker
	// exists for. The tokens are artifacts, so feeding it is ALL the
	// bots can do: see pushLoopToken.
	owner := g.Seats[0].ID
	g.WithWriteLock(func() {
		g.LoopThreshold = threshold
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

	pol := func() aiseat.Policy { return &scripted{prefer: []string{"Keep hand"}} }
	r0 := aiseat.Start(ctx, room, g.Seats[0].ID, pol(), aiseat.Config{}, nil, testLogger())
	r1 := aiseat.Start(ctx, room, g.Seats[1].ID, pol(), aiseat.Config{}, nil, testLogger())
	waitFor(t, "both seats to keep", func() bool {
		return handKept(g, g.Seats[0].ID) && handKept(g, g.Seats[1].ID)
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
	//
	// The wait has to describe the state the ASSERTIONS need, not a
	// state the runners pass through on the way to it. That is #924.
	// "Both parked and the notice up" is already true at the instant
	// the breaker first fires — with the loop_shortcut prompt sitting
	// in the queue, unanswered, and the tally still at the threshold —
	// because AutoPassSuspended goes true with the notice and a runner
	// reports Idle in the window between waking and dispatching. The
	// test then read a table that had not finished settling, and the
	// "1 prompt still open" and "resolved 5 times, want 15" lines were
	// two names for that one instant.
	//
	// So the wait asks for the whole of it: both seats parked with no
	// wake in hand, the notice standing, nothing owed, and the
	// shortcut already spent. The three game-side facts come out of
	// ONE read snapshot, because three facts read one after another
	// are three facts about three different tables — the same rule
	// that governs every other wait in this package (#848).
	//
	// `>=` on the count, not `==`: a shortcut that ran for longer than
	// the bot agreed to has to reach the assertion below and fail
	// there, naming the number, rather than spin here until the budget
	// runs out. And the state this waits for is terminal by
	// construction — parked bots holding on a pass they will not make,
	// with nothing else writing to this room — so it cannot be reached
	// and then left behind.
	wantResolutions := threshold + game.DefaultLoopShortcutIterations
	waitFor(t, "the table to settle: the shortcut spent, both bots parked and nothing owed", func() bool {
		if !r0.Idle() || !r1.Idle() {
			return false
		}
		settled := false
		g.ReadSnapshot(func() {
			// g.LoopNotice, not g.AutoPassSuspended(): the accessor
			// takes the read lock this snapshot is already holding,
			// and a second RLock from the same goroutine deadlocks
			// outright whenever a writer is queued between them —
			// #848's first flake, and a bot runner is exactly that
			// writer.
			settled = g.LoopNotice != nil &&
				owedChoicesLocked(g) == 0 &&
				busiestResolutionsLocked(g) >= wantResolutions
		})
		return settled
	})
	notice := g.CurrentLoopNotice()
	if notice == nil || notice.Count < threshold {
		t.Fatalf("loop notice = %+v, want a count of at least the threshold", notice)
	}
	// Both asks answered: nothing is left owed, which is why the
	// runners could park at all. The wait already required it, so this
	// is the guard it reads as rather than the thing that catches a
	// race.
	var owed int
	g.ReadSnapshot(func() { owed = owedChoicesLocked(g) })
	if owed != 0 {
		t.Errorf("%d prompts still open with both bots parked — a seat is sitting on a question", owed)
	}

	// The count is the assertion that the shortcut ran once and only
	// once: the threshold's worth of iterations that raised the first
	// notice, plus the K the bot agreed to, and then nothing.
	var got int
	g.ReadSnapshot(func() { got = busiestResolutionsLocked(g) })
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
