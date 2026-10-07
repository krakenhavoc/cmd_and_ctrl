package game

import (
	"testing"

	"github.com/google/uuid"
)

// trigger_batch_test.go — #2450, ADR 0055's amendment of 2026-10-07,
// option A: a batch of triggers that drains the stack is not a loop.
//
// CR 603.2c lets one event trigger an ability many times. Thirty
// creatures dying under Syr Konrad is thirty triggers with no decision
// between them, and before this the breaker called it a loop at the
// threshold and stopped a bot-only table. A loop refills the stack and
// a batch empties it; the engine now tells them apart by the lowest
// amount of work left waiting, measured as each triggered ability
// begins to resolve.

const (
	batchOracle = "test-batch-watcher"
	batchLabel  = "Batch Watcher — note a Plant"
	batchToken  = "Plant"
)

// installBatchWatcher puts a permanent on the battlefield whose ability
// triggers, and does nothing, each time a Plant enters. It never feeds
// itself, so any number of Plants entering at once is a finite batch.
func installBatchWatcher(t *testing.T, g *Game, owner *Player) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.Battlefield.PushTop(Card{InstanceID: id, Name: "Batch Watcher", OracleID: batchOracle,
		TypeLine: "Enchantment", Owner: owner.ID, Controller: owner.ID})
	withCatalogTriggers(t, func(oracle string) []TriggeredAbility {
		if oracle != batchOracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventETB},
			AppliesTo: func(ev Event, _ *Card, _ Characteristic, g *Game) bool {
				c := g.findCardByIDLocked(ev.CardID)
				return c != nil && c.Name == batchToken
			},
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return newTriggeredItemForTest(source, batchLabel, func(*Game, *StackItem) error { return nil })
			},
		}}
	})
	return id
}

// passBatchUntilEmpty passes around the table until nothing is left on
// the stack or waiting, failing if the loop breaker fires on the way.
func passBatchUntilEmpty(t *testing.T, g *Game) {
	t.Helper()
	for i := 0; i < 1000; i++ {
		if g.LoopNotice != nil {
			t.Fatalf("the loop breaker fired on a draining batch after %d passes: %+v", i, g.LoopNotice)
		}
		if !g.stackHasItemsLocked() && len(g.PendingTriggers) == 0 {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never emptied")
}

// The headline: ten triggers of one ability from one batch of Plants,
// against a threshold of three. They all resolve and nothing fires.
func TestADrainingBatchDoesNotFireTheLoopBreaker(t *testing.T) {
	const threshold, plants = 3, 10
	g := newActiveGame(t)
	g.LoopThreshold = threshold
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	src := installBatchWatcher(t, g, p)

	g.WithWriteLock(func() {
		for i := 0; i < plants; i++ {
			g.pushLoopToken(batchToken, p.ID, p.ID)
		}
		g.runStateChecksLocked()
	})
	passBatchUntilEmpty(t, g)

	if got := g.TurnTally.Resolved[g.objectTallyKeyLocked(src, batchLabel)]; got != plants {
		t.Errorf("the watcher resolved %d times, want %d", got, plants)
	}
	if n := countLoopEvents(g); n != 0 {
		t.Errorf("EventLoopSuspected count = %d, want 0", n)
	}
}

// The other half: a real loop never makes a new low, so it still
// trips at the threshold. This is #628's token pair under the new
// rule.
func TestARealTriggerLoopStillFiresWithTheBatchRule(t *testing.T) {
	const threshold = 4
	g := newActiveGame(t)
	g.LoopThreshold = threshold
	seedTriggerLoop(t, g)
	passUntilLoopNotice(t, g)
	if g.LoopNotice.Count != threshold {
		t.Errorf("notice count = %d, want %d", g.LoopNotice.Count, threshold)
	}
}

// #810 is untouched: a free activation whose every resolution sets off
// a draining batch is still a loop of activations, and still trips.
// Without TurnTally.LoopActivated the batch's new lows would wipe the
// activation's run every time round.
func TestABatchBetweenActivationsDoesNotHideAnActivationLoop(t *testing.T) {
	const threshold = 4
	g := newActiveGame(t)
	g.LoopThreshold = threshold
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	installBatchWatcher(t, g, p)

	c := NewCard("Plant Press", p.ID)
	c.TypeLine = "Enchantment"
	c.Controller = p.ID
	const pressLabel = "Plant Press: make three Plants"
	c.ActivatedAbilities = []ActivatedAbilityShape{{
		Label: pressLabel,
		Effect: func(g *Game, _ *StackItem) error {
			for i := 0; i < 3; i++ {
				g.pushLoopToken(batchToken, p.ID, p.ID)
			}
			return nil
		},
	}}
	g.Battlefield.PushTop(c)

	for i := 0; i < threshold && g.LoopNotice == nil; i++ {
		if err := g.ActivateCatalogAbility(p.ID, c.InstanceID, 0, ActivateAbilityParams{}); err != nil {
			t.Fatalf("activation %d: %v", i+1, err)
		}
		for j := 0; j < 50 && g.LoopNotice == nil && (g.stackHasItemsLocked() || len(g.PendingTriggers) > 0); j++ {
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
		}
	}
	if g.LoopNotice == nil {
		t.Fatalf("no notice after %d activations of %q, each followed by a batch of three", threshold, pressLabel)
	}
	if g.LoopNotice.Source != c.InstanceID || g.LoopNotice.Label != pressLabel {
		t.Errorf("notice = %s %q, want the Plant Press", g.LoopNotice.Source, g.LoopNotice.Label)
	}
}

// Owner decision 5: a new low while a notice stands clears it and
// withdraws the shortcut prompt, because the notice's claim no longer
// holds.
func TestANewLowClearsAStandingNotice(t *testing.T) {
	g := newActiveGame(t)
	g.LoopThreshold = 3
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	src := installBatchWatcher(t, g, p)

	g.WithWriteLock(func() {
		for i := 0; i < 4; i++ {
			g.pushLoopToken(batchToken, p.ID, p.ID)
		}
		g.runStateChecksLocked()
		// Stand a notice up by hand, as if the breaker had fired on
		// this ability, with its prompt queued.
		g.TurnTally.LoopRun = map[string]int{TallyKey(src, batchLabel): 3}
		g.LoopNotice = &LoopNotice{Source: src, Label: batchLabel, Controller: p.ID, Count: 3}
		g.queueLoopShortcutLocked(TallyKey(src, batchLabel), false)
	})
	if loopShortcutPrompt(g) == nil {
		t.Fatal("setup: no shortcut prompt")
	}
	g.WithWriteLock(func() {
		g.TurnTally.LoopLow, g.TurnTally.LoopLowSet = 99, true
		g.noteLoopProgressLocked()
	})
	if g.LoopNotice != nil {
		t.Errorf("notice still up after progress: %+v", g.LoopNotice)
	}
	if loopShortcutPrompt(g) != nil {
		t.Error("the shortcut prompt was left behind, and it blocks the table")
	}
	if len(g.TurnTally.LoopRun) != 0 {
		t.Errorf("runs not restarted: %v", g.TurnTally.LoopRun)
	}
}
