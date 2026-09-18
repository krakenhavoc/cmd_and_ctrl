package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// loop_breaker_test.go drives a REAL two-permanent trigger loop
// through the real priority loop (#628, CR 726).
//
// The pair is the one the issue describes: two permanents that each
// make a token off the other's token entering. Seeding one token
// starts it, and from then on every pass around the table resolves
// one trigger and queues the next, forever, exactly as four browsers
// on autopass would drive it. The engine test stands in for those
// browsers by calling PassPriority itself — autopass is client code
// (ADR 0009), and what the client reads is the state these tests
// assert on.

const (
	loopOracleMirror = "test-loop-mirror"
	loopOracleSpark  = "test-loop-spark"
	loopLabelMirror  = "Mirror Engine — create a Spark"
	loopLabelSpark   = "Spark Engine — create a Mirror Shard"
	loopTokenSpark   = "Spark"
	loopTokenMirror  = "Mirror Shard"
)

// loopTrigger builds the synthetic ability: when a token named
// `watch` enters, create one named `make`.
func loopTrigger(watch, label, make string) TriggeredAbility {
	return TriggeredAbility{
		Watches: []EventKind{EventETB},
		AppliesTo: func(ev Event, _ *Card, _ Characteristic, g *Game) bool {
			c := g.findCardByIDLocked(ev.CardID)
			return c != nil && c.Name == watch
		},
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			// Capture the controller by value: `source` points into a
			// zone slice that may have moved by resolution time.
			controller, owner := source.Controller, source.Owner
			return NewTriggeredItem(source, label, func(g *Game, _ *StackItem) error {
				g.pushLoopToken(make, controller, owner)
				return nil
			})
		},
	}
}

// pushLoopToken puts one synthetic token on the battlefield and
// announces it, which is all the loop needs to keep going. Caller
// must hold g.mu.
func (g *Game) pushLoopToken(name string, controller, owner uuid.UUID) {
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Token Creature — Elemental",
		Power:      1,
		Toughness:  1,
		Owner:      owner,
		Controller: controller,
	})
	g.EmitEvent(Event{Kind: EventETB, CardID: id, Actor: controller})
}

// seedTriggerLoop puts the two engines on the battlefield, installs
// their triggers, and drops the first Spark so the loop is already
// turning — one trigger on the stack, waiting to be passed onto.
// Returns the two permanents' instance IDs.
func seedTriggerLoop(t *testing.T, g *Game) (mirror, spark uuid.UUID) {
	t.Helper()
	p := g.Seats[0]
	mirror, spark = uuid.New(), uuid.New()
	for _, c := range []Card{
		{InstanceID: mirror, Name: "Mirror Engine", OracleID: loopOracleMirror,
			TypeLine: "Artifact", Owner: p.ID, Controller: p.ID},
		{InstanceID: spark, Name: "Spark Engine", OracleID: loopOracleSpark,
			TypeLine: "Artifact", Owner: p.ID, Controller: p.ID},
	} {
		g.Battlefield.PushTop(c)
	}
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		switch id {
		case loopOracleMirror:
			return []TriggeredAbility{loopTrigger(loopTokenMirror, loopLabelMirror, loopTokenSpark)}
		case loopOracleSpark:
			return []TriggeredAbility{loopTrigger(loopTokenSpark, loopLabelSpark, loopTokenMirror)}
		}
		return nil
	})
	g.WithWriteLock(func() {
		g.pushLoopToken(loopTokenSpark, p.ID, p.ID)
		g.runStateChecksLocked()
	})
	if len(g.StackMeta) != 1 {
		t.Fatalf("seed: stack items = %d, want 1 (the first trigger)", len(g.StackMeta))
	}
	return mirror, spark
}

// passAroundOnce passes priority once per seat, which resolves the
// top of the stack and drains the trigger the resolution queued.
func passAroundOnce(t *testing.T, g *Game) {
	t.Helper()
	for i := 0; i < len(g.Seats); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
}

// passUntilLoopNotice passes around the table until the breaker fires,
// and stops there. Since #804 it has to stop there: the notice comes
// with a CR 726 shortcut prompt, and that prompt blocks the table
// until the loop's controller answers it.
func passUntilLoopNotice(t *testing.T, g *Game) {
	t.Helper()
	for i := 0; i < 200; i++ {
		if g.LoopNotice != nil {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority while winding the loop up: %v", err)
		}
	}
	t.Fatal("the loop breaker never fired")
}

// loopShortcutPrompt returns the outstanding CR 726 prompt, or nil.
func loopShortcutPrompt(g *Game) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceLoopShortcut {
			return c
		}
	}
	return nil
}

// answerLoopShortcut answers the outstanding CR 726 prompt with K.
func answerLoopShortcut(t *testing.T, g *Game, iterations int) {
	t.Helper()
	c := loopShortcutPrompt(g)
	if c == nil {
		t.Fatal("no CR 726 shortcut prompt to answer")
	}
	if err := g.ResolveLoopShortcut(c.ID, c.Chooser, iterations); err != nil {
		t.Fatalf("ResolveLoopShortcut(%d): %v", iterations, err)
	}
}

func countLoopEvents(g *Game) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == EventLoopSuspected {
			n++
		}
	}
	return n
}

// TestLoopBreakerFiresAtTheThreshold is the headline: the notice
// appears on the Nth resolution of ONE ability and not before.
//
// The two engines alternate, so the Spark engine's ability (which
// goes first) reaches N on total resolution 2N-1. The test asserts
// the boundary from both sides — silent at N-1, raised at N — which
// is the assertion that would catch an off-by-one making the breaker
// fire a turn early in real play.
func TestLoopBreakerFiresAtTheThreshold(t *testing.T) {
	const threshold = 4
	g := newActiveGame(t)
	g.LoopThreshold = threshold
	_, spark := seedTriggerLoop(t, g)

	for i := 1; i < threshold; i++ {
		passAroundOnce(t, g) // Spark engine resolves
		passAroundOnce(t, g) // Mirror engine resolves
		if g.LoopNotice != nil {
			t.Fatalf("notice raised after %d resolutions of each ability, want silence below %d", i, threshold)
		}
	}
	if got := g.TurnTally.LoopRun[TallyKey(spark, loopLabelSpark)]; got != threshold-1 {
		t.Fatalf("loop run before the threshold = %d, want %d", got, threshold-1)
	}

	passAroundOnce(t, g)
	if g.LoopNotice == nil {
		t.Fatalf("no loop notice after %d resolutions of %q", threshold, loopLabelSpark)
	}
	if g.LoopNotice.Source != spark {
		t.Errorf("notice source = %s, want the Spark engine %s", g.LoopNotice.Source, spark)
	}
	if g.LoopNotice.Label != loopLabelSpark {
		t.Errorf("notice label = %q, want %q", g.LoopNotice.Label, loopLabelSpark)
	}
	if g.LoopNotice.Count != threshold {
		t.Errorf("notice count = %d, want %d", g.LoopNotice.Count, threshold)
	}
	if g.LoopNotice.Controller != g.Seats[0].ID {
		t.Errorf("notice controller = %s, want %s", g.LoopNotice.Controller, g.Seats[0].ID)
	}
	if !g.AutoPassSuspended() {
		t.Error("AutoPassSuspended() = false with a notice standing")
	}
	if n := countLoopEvents(g); n != 1 {
		t.Errorf("EventLoopSuspected count = %d, want exactly 1 per run", n)
	}
}

// TestLoopBreakerStopsAutomaticPassingNotTheGame pins the half of the
// rule that is easy to get wrong: the engine suspends AUTOMATIC
// passing, it does not refuse a pass. A table that wants to watch the
// loop run says so one click at a time, and the notice stays up while
// it does — a bare pass is not a decision.
func TestLoopBreakerStopsAutomaticPassingNotTheGame(t *testing.T) {
	const threshold = 3
	g := newActiveGame(t)
	g.LoopThreshold = threshold
	seedTriggerLoop(t, g)
	passUntilLoopNotice(t, g)
	if g.LoopNotice.Label != loopLabelSpark {
		t.Fatalf("notice label = %q, want the ability that tripped first (%q)", g.LoopNotice.Label, loopLabelSpark)
	}
	// #804: the shortcut prompt is the one thing that DOES hold the
	// table, and only until it is answered. See ADR 0055's 2026-09-18
	// amendment for why this prompt blocks where the notice does not.
	if err := g.PassPriority(); !errors.Is(err, ErrChoicePending) {
		t.Fatalf("PassPriority with the CR 726 prompt open = %v, want ErrChoicePending", err)
	}
	answerLoopShortcut(t, g, 0)
	if g.LoopNotice == nil {
		t.Fatal("answering stop-here cleared the notice; the table should stay paused")
	}
	before := g.LoopNotice.Count

	// Every pass is still accepted, and the loop still resolves: two
	// more passes around the table bring the tripping ability back.
	passAroundOnce(t, g)
	passAroundOnce(t, g)
	if g.LoopNotice == nil {
		t.Fatal("a bare pass cleared the notice; only a decision should")
	}
	if g.LoopNotice.Count <= before {
		t.Errorf("notice count = %d after another resolution, want > %d", g.LoopNotice.Count, before)
	}
	if g.LoopNotice.Label != loopLabelSpark {
		t.Errorf("notice label = %q; the loop's other half replaced it", g.LoopNotice.Label)
	}
	if n := countLoopEvents(g); n != 1 {
		t.Errorf("EventLoopSuspected count = %d, want 1 — the breadcrumb is once per run", n)
	}
}

// TestAnsweredPromptClearsTheLoopNotice — answering a prompt is a
// decision, and a decision is what "until a human acts" means. The
// run restarts from zero, so the breaker will not fire again until
// the loop has run the full threshold afresh.
func TestAnsweredPromptClearsTheLoopNotice(t *testing.T) {
	const threshold = 3
	g := newActiveGame(t)
	g.LoopThreshold = threshold
	seedTriggerLoop(t, g)
	passUntilLoopNotice(t, g)

	p := g.Seats[0]
	var choiceID uuid.UUID
	g.WithWriteLock(func() {
		choiceID = g.QueueChoiceForEffect(PendingChoice{
			Kind:    PendingChoiceConfirm,
			Chooser: p.ID,
			Reason:  "Mirror Engine — keep going?",
		})
	})
	if err := g.ResolveConfirm(choiceID, p.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}

	if g.LoopNotice != nil {
		t.Errorf("notice survived an answered prompt: %+v", g.LoopNotice)
	}
	if g.AutoPassSuspended() {
		t.Error("AutoPassSuspended() = true after a player decision")
	}
	if len(g.TurnTally.LoopRun) != 0 {
		t.Errorf("loop run = %v after a decision, want empty", g.TurnTally.LoopRun)
	}
	// #804: the shortcut prompt goes with the notice it was asking
	// about. It blocks the table, so a stale one is not a stale
	// question — it is a game nobody can move on.
	if c := loopShortcutPrompt(g); c != nil {
		t.Errorf("the CR 726 prompt outlived its notice: %+v", c)
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Errorf("AdvanceStep after the decision: %v", err)
	}
}

// TestDecisionEventsRestartTheRun covers the notches the tally
// listener owns — a cast, a declared attacker, a declared blocker —
// at the level they are wired. Each one restarts the run without
// touching TurnTally.Resolved, which other cards read for their own
// once-per-turn gates and which this feature must not disturb.
func TestDecisionEventsRestartTheRun(t *testing.T) {
	const label = "X — again"
	for _, tc := range []struct {
		name string
		kind EventKind
	}{
		{"cast", EventCast},
		{"block declared", EventBlock},
		{"attack declared", EventAttack},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			g.LoopThreshold = 2
			p := g.Seats[0]
			source := uuid.New()
			g.WithWriteLock(func() {
				g.EmitEvent(Event{Kind: EventResolve, Actor: p.ID, Source: source, Label: label})
				g.EmitEvent(Event{Kind: tc.kind, Actor: p.ID, CardID: uuid.New()})
				g.EmitEvent(Event{Kind: EventResolve, Actor: p.ID, Source: source, Label: label})
			})
			if g.LoopNotice != nil {
				t.Errorf("breaker fired across a %s: %+v", tc.name, g.LoopNotice)
			}
			if got := g.TurnTally.LoopRun[TallyKey(source, label)]; got != 1 {
				t.Errorf("loop run = %d after a %s, want 1", got, tc.name)
			}
			if got := g.TurnTally.Resolved[TallyKey(source, label)]; got != 2 {
				t.Errorf("Resolved = %d, want 2 — a decision restarts the run, not the turn tally", got)
			}
		})
	}
}

// TestOrdinaryRepeatedTriggersDoNotFireTheBreaker is the
// false-positive guard, and the reason the count is keyed by
// TallyKey(source, label) rather than by "triggers this turn": four
// upkeep triggers from four players are four different keys. Here
// four sources share ONE label and each resolves twice — eight
// resolutions in one turn against a threshold of three — and nothing
// fires, because no single ability got there.
func TestOrdinaryRepeatedTriggersDoNotFireTheBreaker(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	g.LoopThreshold = 3
	const label = "Upkeep Drum — draw a card"

	g.WithWriteLock(func() {
		for round := 0; round < 2; round++ {
			for _, p := range g.Seats {
				g.EmitEvent(Event{
					Kind:   EventResolve,
					Actor:  p.ID,
					Source: uuid.NewSHA1(uuid.Nil, []byte(p.ID.String())),
					Label:  label,
				})
			}
		}
	})

	if g.LoopNotice != nil {
		t.Fatalf("breaker fired on ordinary per-player triggers: %+v", g.LoopNotice)
	}
	if n := countLoopEvents(g); n != 0 {
		t.Fatalf("EventLoopSuspected count = %d, want 0", n)
	}
	// Eight resolutions counted, none of them the same ability more
	// than twice.
	for k, n := range g.TurnTally.LoopRun {
		if n >= 3 {
			t.Errorf("run for %q = %d, want < 3", k, n)
		}
	}
}

// TestLoopNoticeSurvivesCloneAndSnapshot — the notice is per-turn
// state an undo must be able to rewind past and a restore must not
// silently drop, so it rides both copies like every other field on
// Game (snapshot_drift_test.go classifies it `carried`).
func TestLoopNoticeSurvivesCloneAndSnapshot(t *testing.T) {
	const threshold = 3
	g := newActiveGame(t)
	g.LoopThreshold = threshold
	seedTriggerLoop(t, g)
	passUntilLoopNotice(t, g)
	want := *g.LoopNotice

	clone := g.Clone()
	if clone.LoopNotice == nil || *clone.LoopNotice != want {
		t.Errorf("clone notice = %+v, want %+v", clone.LoopNotice, want)
	}
	if clone.LoopNotice == g.LoopNotice {
		t.Error("clone shares the live notice pointer; a rewind would mutate the original")
	}
	if clone.LoopThreshold != threshold {
		t.Errorf("clone threshold = %d, want %d", clone.LoopThreshold, threshold)
	}

	restored, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if restored.LoopNotice == nil || *restored.LoopNotice != want {
		t.Errorf("restored notice = %+v, want %+v", restored.LoopNotice, want)
	}
	if restored.LoopThreshold != threshold {
		t.Errorf("restored threshold = %d, want %d", restored.LoopThreshold, threshold)
	}
}
