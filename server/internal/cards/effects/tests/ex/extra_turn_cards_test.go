package testsex

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestTimeWarpGivesTheTargetPlayerTheNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	opp := (me + 2) % len(g.Seats)
	effectstest.CastCatalogSpell(t, g, "Time Warp", "Sorcery", effectstest.TimeWarpOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[opp].ID}})
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, opp, true, "after Time Warp's turn")
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, (me+1)%len(g.Seats), false, "after the extra turn (rotation resumes after the caster)")
}

func TestTimeStretchGivesTwoTurns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	effectstest.CastCatalogSpell(t, g, "Time Stretch", "Sorcery", effectstest.TimeStretchOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[me].ID}})
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, me, true, "first Time Stretch turn")
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, me, true, "second Time Stretch turn")
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, (me+1)%len(g.Seats), false, "after both extra turns")
}

func TestTakeAnExtraTurnSorceries(t *testing.T) {
	for _, c := range []struct{ name, oracle string }{
		{"Temporal Manipulation", effectstest.TemporalManipulationOracle},
		{"Capture of Jingzhou", effectstest.CaptureOfJingzhouOracle},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Turn.ActiveSeat
			spell := effectstest.CastCatalogSpell(t, g, c.name, "Sorcery", c.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Seats[me].Graveyard.Contains(spell) {
				t.Errorf("%s should be in its owner's graveyard", c.name)
			}
			effectstest.EndTurn(t, g)
			effectstest.AssertTurnOf(t, g, me, true, "after "+c.name)
		})
	}
}

// Temporal Mastery exiles itself, so it is not left in the graveyard.
func TestTemporalMasteryTakesATurnAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	spell := effectstest.CastCatalogSpell(t, g, "Temporal Mastery", "Sorcery", effectstest.TemporalMasteryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Seats[me].Graveyard.Contains(spell) || !g.Exile.Contains(spell) {
		t.Errorf("Temporal Mastery should be in exile (graveyard %v, exile %v)",
			g.Seats[me].Graveyard.Contains(spell), g.Exile.Contains(spell))
	}
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, me, true, "after Temporal Mastery")
}

// Final Fortune's loss fires in the EXTRA turn's end step: not in the
// end step of the turn it was cast in, which an unbound "next end step"
// trigger would have picked, and then the player is out.
func TestFinalFortuneLosesAtTheExtraTurnsEndStep(t *testing.T) {
	for _, c := range []struct{ name, typeLine, oracle string }{
		{"Final Fortune", "Instant", effectstest.FinalFortuneOracle},
		{"Last Chance", "Sorcery", effectstest.LastChanceOracle},
		{"Warrior's Oath", "Sorcery", effectstest.WarriorsOathOracle},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Turn.ActiveSeat
			effectstest.CastCatalogSpell(t, g, c.name, c.typeLine, c.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)

			effectstest.AdvanceTo(t, g, game.StepEnd)
			effectstest.PassPriorityAroundTable(t, g)
			if g.Seats[me].Eliminated {
				t.Fatal("lost in the end step of the turn the spell resolved in")
			}
			if len(g.DelayedTriggers) != 1 {
				t.Fatalf("the loss trigger should still be waiting: %d queued", len(g.DelayedTriggers))
			}

			effectstest.EndTurn(t, g)
			effectstest.AssertTurnOf(t, g, me, true, "the extra turn")
			effectstest.AdvanceTo(t, g, game.StepEnd)
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Seats[me].Eliminated {
				t.Fatal("did not lose at the beginning of the extra turn's end step")
			}
			if g.State != game.StateActive {
				t.Fatalf("three players remain; the game should go on: %s", g.State)
			}
			effectstest.AssertTurnOf(t, g, (me+1)%len(g.Seats), false, "after the loser's extra turn")
		})
	}
}

// A second extra turn taken during Final Fortune's turn comes AFTER
// it (CR 500.7 only reorders turns not yet begun), so it does not
// rescue the player; the loss still fires in the first extra turn.
func TestFinalFortuneIsNotRescuedByALaterExtraTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	effectstest.CastCatalogSpell(t, g, "Final Fortune", "Instant", effectstest.FinalFortuneOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.EndTurn(t, g)
	effectstest.CastCatalogSpell(t, g, "Temporal Manipulation", "Sorcery", effectstest.TemporalManipulationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Seats[me].Eliminated {
		t.Fatal("the later extra turn should not have delayed Final Fortune's loss")
	}
}

func TestMagistratesScepterChargesAndTakesATurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	scepter := effectstest.B12Push(g, me.ID, "Magistrate's Scepter", "Artifact", effectstest.MagistratesScepterOracle, 0, 0)
	effectstest.ToMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, scepter, 1, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the extra-turn ability activated without three charge counters")
	}
	if err := g.ActivateCatalogAbility(me.ID, scepter, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("charge: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, scepter, game.CounterCharge); got != 1 {
		t.Fatalf("charge counters %d, want 1", got)
	}

	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if c := &g.Battlefield.Cards[i]; c.InstanceID == scepter {
				c.Counters[game.CounterCharge] = 3
				c.Tapped = false
			}
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, scepter, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("take a turn: %v", err)
	}
	if got := effectstest.CounterCount(g, scepter, game.CounterCharge); got != 0 {
		t.Errorf("charge counters %d after paying the cost, want 0", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, me.Seat, true, "after the Scepter")
}

func TestTeferiMasterOfTimeMinusTenTakesTwoTurns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	teferi := effectstest.PushCatalogWalker(g, me.ID, "Teferi, Master of Time", effectstest.TeferiMasterOfTimeOracle, 10)
	effectstest.ToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, teferi, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−10: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, me.Seat, true, "first −10 turn")
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, me.Seat, true, "second −10 turn")
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, (me.Seat+1)%len(g.Seats), false, "after both turns")
}

// Avatar Kuruk's exhaust ability takes a turn, and only once.
func TestAvatarKurukExhaustTakesAnExtraTurnOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kuruk := effectstest.B12Push(g, me.ID, "Avatar Kuruk", "Legendary Creature — Avatar", effects.TheLegendOfKurukOracleID+"#1", 4, 3)
	effectstest.ToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, kuruk, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Waterbend {20}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, kuruk, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("an exhaust ability activated twice")
	}
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, me.Seat, true, "after Avatar Kuruk's ability")
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, (me.Seat+1)%len(g.Seats), false, "after the extra turn")
}
