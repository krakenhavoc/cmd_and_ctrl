package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lki_copy_test.go — #1255, CR 608.2h against CR 608.2b.
//
// A copy effect that NAMES a spell without targeting it copies the
// spell from last-known information once it has left the stack; one
// that TARGETS the spell does nothing, because its only target is
// illegal. Storm is pinned in storm_test.go. This file holds the other
// two non-targeting families that opted in (Thousand-Year Storm and
// Doublecast's "copy that spell") and the targeting family that must
// NOT have changed (Reverberate).

// counterNow counters a spell on the stack, as a counterspell cast in
// response would.
func counterNow(t *testing.T, g *game.Game, spell uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(spell); err != nil {
			t.Fatalf("CounterTargetForEffect: %v", err)
		}
	})
	if g.Stack.Contains(spell) {
		t.Fatal("the spell was not countered — the fixture is wrong")
	}
}

// settleKeepingTarget passes priority until the stack is empty,
// answering every copy's re-target prompt with `victim` and every
// trigger-order prompt as offered. Returns how many re-target prompts
// it answered.
func settleKeepingTarget(t *testing.T, g *game.Game, chooser, victim uuid.UUID) int {
	t.Helper()
	answered := 0
	for i := 0; i < 64; i++ {
		if p := triggerOrderPromptFor(g, chooser); p != nil {
			if err := g.ResolveTriggerOrder(p.ID, chooser, append([]uuid.UUID(nil), p.TriggerOrderIDs...)); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if p := latestPickTarget(g, chooser); p != nil {
			if err := g.ResolvePickTarget(p.ID, chooser,
				game.TargetRef{Kind: game.TargetPlayer, ID: victim}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			answered++
			continue
		}
		if stackFullyEmpty(g) {
			return answered
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never emptied")
	return answered
}

// TestThousandYearStormCopiesACounteredSpell — the second instant of
// the turn triggers Thousand-Year Storm for one copy; the Bolt is
// countered in response to the trigger, and the copy is still made.
func TestThousandYearStormCopiesACounteredSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Thousand-Year Storm", "Enchantment", b12ThousandYearStormOracle, 0, 0)
	before := opp.Life

	b12Bolt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	counterNow(t, g, bolt)

	if got := settleKeepingTarget(t, g, me.ID, opp.ID); got != 1 {
		t.Errorf("%d re-target prompts, want 1 (one copy)", got)
	}
	if opp.Life != before-6 {
		t.Errorf("life %d → %d, want -6: the first Bolt and the countered second Bolt's copy (CR 608.2h)", before, opp.Life)
	}
}

// TestDoublecastCopiesTheNextInstantEvenIfItIsCountered — "copy that
// spell" names the spell the delayed trigger saw cast; it does not
// target it.
func TestDoublecastCopiesTheNextInstantEvenIfItIsCountered(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0].ID, g.Seats[1].ID
	before := lifeOf(g, victim)

	castCatalogSpell(t, g, "Doublecast", "Sorcery", doublecastOracle, nil)
	passPriorityAroundTable(t, g)
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim}})
	counterNow(t, g, bolt)

	if got := settleKeepingTarget(t, g, me, victim); got != 1 {
		t.Errorf("%d re-target prompts, want 1 (one copy)", got)
	}
	if got := lifeOf(g, victim); got != before-3 {
		t.Errorf("life %d → %d, want -3: the countered Bolt deals nothing, its copy deals 3", before, got)
	}
}

// TestReverberateOnACounteredSpellCopiesNothing — the regression the
// whole change is shaped around. Reverberate TARGETS the spell, so
// when the Bolt is countered in response Reverberate's only target is
// illegal and CR 608.2b counters it by game rules: no copy, no prompt,
// no damage. The last-known record now exists for the Bolt, and
// Reverberate must not read it.
func TestReverberateOnACounteredSpellCopiesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0].ID, g.Seats[1].ID
	before := lifeOf(g, victim)

	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim}})
	castCatalogSpell(t, g, "Reverberate", "Instant", reverberateOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	counterNow(t, g, bolt)

	if got := settleKeepingTarget(t, g, me, victim); got != 0 {
		t.Errorf("%d re-target prompts, want 0 — Reverberate's target is gone (CR 608.2b)", got)
	}
	if got := lifeOf(g, victim); got != before {
		t.Errorf("life %d → %d, want unchanged: the Bolt was countered and Reverberate copies nothing", before, got)
	}
}

// TestCopySpellWithoutLastKnownStillRefusesACounteredSpell — the
// primitive's default is the strict lookup, pinned directly so that a
// future caller cannot reach the record by forgetting the flag's
// meaning: a targeting-style CopySpell on a countered spell is a no-op,
// and the same call with FromLastKnown makes the copy.
func TestCopySpellWithoutLastKnownStillRefusesACounteredSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0].ID, g.Seats[1].ID

	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim}})
	counterNow(t, g, bolt)

	stackSize := func() int { return g.Stack.Size() }
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: me})
		if err := (CopySpell{StackID: bolt, Controller: me}).Apply(ctx); err != nil {
			t.Fatalf("strict CopySpell: %v", err)
		}
	})
	if n := stackSize(); n != 0 {
		t.Fatalf("the strict copy put %d objects on the stack, want 0", n)
	}
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: me})
		if err := (CopySpell{StackID: bolt, Controller: me, FromLastKnown: true}).Apply(ctx); err != nil {
			t.Fatalf("last-known CopySpell: %v", err)
		}
	})
	if n := stackSize(); n != 1 {
		t.Fatalf("the last-known copy put %d objects on the stack, want 1", n)
	}
}
