package teststo

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestTroubleInPairsSkipsAnOpponentsExtraTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Turn.ActiveSeat
	other := (caster + 1) % len(g.Seats)
	effectstest.SeedEnchantment(g, effectstest.TroubleInPairsOracle, "Trouble in Pairs", "Enchantment", g.Seats[other].ID)
	begun := g.Seats[caster].TurnsBegun

	effectstest.CastCatalogSpell(t, g, "Temporal Manipulation", "Sorcery", effectstest.TemporalManipulationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(g.ExtraTurnsQueued()); got != 1 {
		t.Fatalf("the spell should still queue its turn: %d queued", got)
	}

	effectstest.EndTurn(t, g)
	// The turn never begins: rotation moves on to the next seat's NORMAL
	// turn, and the queue is empty.
	effectstest.AssertTurnOf(t, g, other, false, "after the skipped extra turn")
	if got := len(g.ExtraTurnsQueued()); got != 0 {
		t.Errorf("the skipped turn is still queued: %d", got)
	}
	if got := effectstest.SkippedExtraTurnEvents(g); got != 1 {
		t.Errorf("EventExtraTurnSkipped emitted %d times, want 1", got)
	}
	if got := g.Seats[caster].TurnsBegun; got != begun {
		t.Errorf("a skipped turn counted as begun: TurnsBegun %d, want %d", got, begun)
	}
}

// Time Stretch queues two turns for the same opponent; each is its own
// window, so both are skipped.
func TestTroubleInPairsSkipsEveryTurnOfTimeStretch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Turn.ActiveSeat
	other := (caster + 2) % len(g.Seats)
	effectstest.SeedEnchantment(g, effectstest.TroubleInPairsOracle, "Trouble in Pairs", "Enchantment", g.Seats[other].ID)

	effectstest.CastCatalogSpell(t, g, "Time Stretch", "Sorcery", effectstest.TimeStretchOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[caster].ID}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, (caster+1)%len(g.Seats), false, "after both skipped turns")
	if got := effectstest.SkippedExtraTurnEvents(g); got != 2 {
		t.Errorf("EventExtraTurnSkipped emitted %d times, want 2", got)
	}
}

// A skipped turn takes what was scheduled for it along (CR 614.10a):
// Final Fortune's lose-the-game was bound to that turn's end step and
// must not fire, or be left waiting to fire in some later end step.
func TestTroubleInPairsSkippedTurnTakesFinalFortuneWithIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Turn.ActiveSeat
	other := (caster + 1) % len(g.Seats)
	effectstest.SeedEnchantment(g, effectstest.TroubleInPairsOracle, "Trouble in Pairs", "Enchantment", g.Seats[other].ID)

	effectstest.CastCatalogSpell(t, g, "Final Fortune", "Instant", effectstest.FinalFortuneOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, other, false, "after the skipped turn")
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("Final Fortune's loss is still queued for a turn that will never happen: %d", len(g.DelayedTriggers))
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Seats[caster].Eliminated {
		t.Fatal("the caster lost to Final Fortune although the extra turn was skipped")
	}
}

// "An opponent": the controller's own extra turns are untouched.
func TestTroubleInPairsDoesNotSkipItsControllersExtraTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	effectstest.SeedEnchantment(g, effectstest.TroubleInPairsOracle, "Trouble in Pairs", "Enchantment", g.Seats[me].ID)

	effectstest.CastCatalogSpell(t, g, "Temporal Manipulation", "Sorcery", effectstest.TemporalManipulationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, me, true, "the controller's own extra turn")
	if got := effectstest.SkippedExtraTurnEvents(g); got != 0 {
		t.Errorf("EventExtraTurnSkipped emitted %d times, want 0", got)
	}
}

// The clause lasts as long as the enchantment does (CR 113.6): with it
// gone by the time the turn would begin, the turn is taken.
func TestTroubleInPairsGoneBeforeTheTurnBeginsTakesTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Turn.ActiveSeat
	other := (caster + 1) % len(g.Seats)
	tip := effectstest.SeedEnchantment(g, effectstest.TroubleInPairsOracle, "Trouble in Pairs", "Enchantment", g.Seats[other].ID)

	effectstest.CastCatalogSpell(t, g, "Temporal Manipulation", "Sorcery", effectstest.TemporalManipulationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(tip) })
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, caster, true, "extra turn with the enchantment gone")
}

// Two Troubles in Pairs under different controllers are two
// applicable replacements on one event; both cancel, so there is
// nothing to order and the turn is skipped once.
func TestTwoTroublesInPairsSkipTheTurnOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Turn.ActiveSeat
	effectstest.SeedEnchantment(g, effectstest.TroubleInPairsOracle, "Trouble in Pairs", "Enchantment", g.Seats[(caster+1)%4].ID)
	effectstest.SeedEnchantment(g, effectstest.TroubleInPairsOracle, "Trouble in Pairs", "Enchantment", g.Seats[(caster+2)%4].ID)

	effectstest.CastCatalogSpell(t, g, "Temporal Manipulation", "Sorcery", effectstest.TemporalManipulationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, (caster+1)%4, false, "after the skipped turn")
	if len(g.PendingChoices) != 0 {
		t.Errorf("an ordering prompt was queued for two pure cancels: %d", len(g.PendingChoices))
	}
	if got := effectstest.SkippedExtraTurnEvents(g); got != 1 {
		t.Errorf("EventExtraTurnSkipped emitted %d times, want 1", got)
	}
}

// With nothing skipping extra turns, the baseline still holds — the
// window is a no-op for every other table.
func TestExtraTurnWindowIsInvisibleWithoutAReplacement(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	effectstest.CastCatalogSpell(t, g, "Temporal Manipulation", "Sorcery", effectstest.TemporalManipulationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, me, true, "extra turn with no replacement")
}
