package testsl

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestThousandYearStormCopiesACounteredSpell — the second instant of
// the turn triggers Thousand-Year Storm for one copy; the Bolt is
// countered in response to the trigger, and the copy is still made.
func TestThousandYearStormCopiesACounteredSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Thousand-Year Storm", "Enchantment", effectstest.B12ThousandYearStormOracle, 0, 0)
	before := opp.Life

	effectstest.B12Bolt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.CounterNow(t, g, bolt)

	if got := effectstest.SettleKeepingTarget(t, g, me.ID, opp.ID); got != 1 {
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
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0].ID, g.Seats[1].ID
	before := effectstest.LifeOf(g, victim)

	effectstest.CastCatalogSpell(t, g, "Doublecast", "Sorcery", effectstest.DoublecastOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim}})
	effectstest.CounterNow(t, g, bolt)

	if got := effectstest.SettleKeepingTarget(t, g, me, victim); got != 1 {
		t.Errorf("%d re-target prompts, want 1 (one copy)", got)
	}
	if got := effectstest.LifeOf(g, victim); got != before-3 {
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
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0].ID, g.Seats[1].ID
	before := effectstest.LifeOf(g, victim)

	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim}})
	effectstest.CastCatalogSpell(t, g, "Reverberate", "Instant", effectstest.ReverberateOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	effectstest.CounterNow(t, g, bolt)

	if got := effectstest.SettleKeepingTarget(t, g, me, victim); got != 0 {
		t.Errorf("%d re-target prompts, want 0 — Reverberate's target is gone (CR 608.2b)", got)
	}
	if got := effectstest.LifeOf(g, victim); got != before {
		t.Errorf("life %d → %d, want unchanged: the Bolt was countered and Reverberate copies nothing", before, got)
	}
}

// TestCopySpellWithoutLastKnownStillRefusesACounteredSpell — the
// primitive's default is the strict lookup, pinned directly so that a
// future caller cannot reach the record by forgetting the flag's
// meaning: a targeting-style CopySpell on a countered spell is a no-op,
// and the same call with FromLastKnown makes the copy.
func TestCopySpellWithoutLastKnownStillRefusesACounteredSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0].ID, g.Seats[1].ID

	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim}})
	effectstest.CounterNow(t, g, bolt)

	stackSize := func() int { return g.Stack.Size() }
	g.WithWriteLock(func() {
		ctx := effects.NewContext(g, &game.StackItem{Controller: me})
		if err := (effects.CopySpell{StackID: bolt, Controller: me}).Apply(ctx); err != nil {
			t.Fatalf("strict CopySpell: %v", err)
		}
	})
	if n := stackSize(); n != 0 {
		t.Fatalf("the strict copy put %d objects on the stack, want 0", n)
	}
	g.WithWriteLock(func() {
		ctx := effects.NewContext(g, &game.StackItem{Controller: me})
		if err := (effects.CopySpell{StackID: bolt, Controller: me, FromLastKnown: true}).Apply(ctx); err != nil {
			t.Fatalf("last-known CopySpell: %v", err)
		}
	})
	if n := stackSize(); n != 1 {
		t.Fatalf("the last-known copy put %d objects on the stack, want 1", n)
	}
}
