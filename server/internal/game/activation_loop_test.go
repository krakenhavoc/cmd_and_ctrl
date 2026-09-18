package game

import (
	"testing"

	"github.com/google/uuid"
)

// activation_loop_test.go — #810, the activation-shaped loop.
//
// ADR 0055 built the breaker for a TRIGGER loop, which runs itself:
// each resolution queues the next trigger and the only thing the
// table does is pass, so counting resolutions since the last player
// decision finds it. An activation loop is the other shape. Every
// iteration IS a player decision — somebody activates a free,
// repeatable ability — and `ActivateCatalogAbility` restarted every
// run including that ability's own, so `LoopRun` never got past 1 and
// the breaker never fired. A table of bots took Soothsaying's "{X}:
// Look at the top X cards" at X=0 79,519 times in five minutes (#810).
//
// The enumerator's X rule is the fix for that card and for every
// other whose whole effect is X (internal/legal/x.go). This is the
// belt to its braces: a repeated activation of ONE ability, with
// nothing else decided in between, is a loop like any other.

const freeLoopLabel = "Free Engine: do nothing, repeatedly"

// pushFreeActivationSource puts a permanent on the battlefield with an
// ability that costs nothing and does nothing — the shape that makes
// a loop out of a player decision. A second, equally free ability
// rides along so a test can show that activating something ELSE is
// still a decision that restarts the run.
func pushFreeActivationSource(g *Game, owner *Player) uuid.UUID {
	c := NewCard("Free Engine", owner.ID)
	c.TypeLine = "Enchantment"
	c.Controller = owner.ID
	c.ActivatedAbilities = []ActivatedAbilityShape{
		{
			Label:  freeLoopLabel,
			Effect: func(*Game, *StackItem) error { return nil },
		},
		{
			Label:  "Free Engine: do nothing, differently",
			Effect: func(*Game, *StackItem) error { return nil },
		},
	}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// activateFree activates one of the Free Engine's abilities and passes
// the table around once, which resolves it.
func activateFree(t *testing.T, g *Game, src uuid.UUID, index int) {
	t.Helper()
	if err := g.ActivateCatalogAbility(g.Seats[0].ID, src, index, ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility(%d): %v", index, err)
	}
	passAroundOnce(t, g)
}

// The headline: a free ability activated over and over, with nothing
// else decided in between, trips the breaker on the threshold's
// resolution — and the CR 726 prompt goes to its controller, so a
// bot-only table has an answer rather than a spin.
func TestRepeatedFreeActivationFiresTheLoopBreaker(t *testing.T) {
	const threshold = 4
	g := newActiveGame(t)
	g.LoopThreshold = threshold
	advanceTo(t, g, StepPrecombatMain)
	src := pushFreeActivationSource(g, g.Seats[0])

	for i := 1; i < threshold; i++ {
		activateFree(t, g, src, 0)
		if g.LoopNotice != nil {
			t.Fatalf("notice raised after %d activations, want silence below %d", i, threshold)
		}
		if got := g.TurnTally.LoopRun[TallyKey(src, freeLoopLabel)]; got != i {
			t.Fatalf("loop run after %d activations = %d — the activation is clearing its own run (#810)", i, got)
		}
	}

	activateFree(t, g, src, 0)
	if g.LoopNotice == nil {
		t.Fatalf("no loop notice after %d free activations of %q", threshold, freeLoopLabel)
	}
	if g.LoopNotice.Source != src || g.LoopNotice.Label != freeLoopLabel {
		t.Errorf("notice = %s %q, want the Free Engine's repeating ability", g.LoopNotice.Source, g.LoopNotice.Label)
	}
	if g.LoopNotice.Count != threshold {
		t.Errorf("notice count = %d, want %d", g.LoopNotice.Count, threshold)
	}
	if !g.AutoPassSuspended() {
		t.Error("AutoPassSuspended() = false with a notice standing")
	}
	c := loopShortcutPrompt(g)
	if c == nil {
		t.Fatal("no CR 726 shortcut prompt — a bot-only table has nothing to answer and never stops")
	}
	if c.Chooser != g.Seats[0].ID {
		t.Errorf("prompt chooser = %s, want the ability's controller %s", c.Chooser, g.Seats[0].ID)
	}
}

// The other half of the rule, and the reason it is safe: an
// activation is still a player decision about every OTHER ability.
// Alternating two free abilities never trips the breaker, because each
// activation clears the other's run — which is what keeps a real turn
// that activates several things from looking like a loop.
func TestAlternatingActivationsDoNotFireTheLoopBreaker(t *testing.T) {
	const threshold = 3
	g := newActiveGame(t)
	g.LoopThreshold = threshold
	advanceTo(t, g, StepPrecombatMain)
	src := pushFreeActivationSource(g, g.Seats[0])

	for i := 0; i < threshold*3; i++ {
		activateFree(t, g, src, i%2)
		if g.LoopNotice != nil {
			t.Fatalf("notice raised on alternating activations after %d of them", i+1)
		}
	}
}

// A cast between two activations restarts the run too — the ordinary
// decision rule, unchanged by #810. Without this the breaker would
// fire on a real turn that used one ability repeatedly around real
// plays, which is the false positive ADR 0055 §1 exists to avoid.
func TestACastBetweenActivationsRestartsTheRun(t *testing.T) {
	const threshold = 3
	g := newActiveGame(t)
	g.LoopThreshold = threshold
	advanceTo(t, g, StepPrecombatMain)
	src := pushFreeActivationSource(g, g.Seats[0])

	for i := 1; i < threshold; i++ {
		activateFree(t, g, src, 0)
	}
	if got := g.TurnTally.LoopRun[TallyKey(src, freeLoopLabel)]; got != threshold-1 {
		t.Fatalf("loop run = %d, want %d before the cast", got, threshold-1)
	}
	// EventCast is the decision notch a real cast reaches
	// turnTallyListener with (ADR 0055 §3).
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventCast, Actor: g.Seats[0].ID, Source: uuid.New()})
	})
	if got := g.TurnTally.LoopRun[TallyKey(src, freeLoopLabel)]; got != 0 {
		t.Fatalf("loop run = %d after a cast, want the run restarted", got)
	}
	activateFree(t, g, src, 0)
	if g.LoopNotice != nil {
		t.Error("the breaker fired on a turn with a real play in the middle of it")
	}
}
