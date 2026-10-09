package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// extra_turn_cards_test.go — the CR 500.7 first wave (ADR 0059 "Card
// first wave", #753). Each card is driven through the real cast or
// activation and the real rotation seam: the extra turn is asserted by
// who is active after the current turn ends, never by reading the
// queue alone.

const (
	timeWarpOracle             = "dbd6a94b-62ff-4a10-9d52-bdd90b26e425"
	timeStretchOracle          = "72e56963-a9dd-44dc-a4d3-992b4d89dd28"
	temporalManipulationOracle = "6b6ac99a-7548-4cad-9fe5-ea611618ab9e"
	captureOfJingzhouOracle    = "89aa65d9-2502-40b0-90b6-b25a8e9f6155"
	temporalMasteryOracle      = "5c58b8e6-c572-461e-893e-a8c05f20ba17"
	finalFortuneOracle         = "8d0adc5c-3fdd-4e22-b783-5651f8e57b65"
	lastChanceOracle           = "360039a5-1cbd-4ee3-8f94-21b5348e106a"
	warriorsOathOracle         = "574044cf-2e2f-4b5c-b4d1-cf05ba814ab2"
	magistratesScepterOracle   = "487485a6-cf79-48c0-bea9-0ec6b3ee253c"
)

// endTurn ends the current turn through the rotation seam. The sandbox
// pass_turn verb is used rather than walking cleanup, because the
// catalog harness's starting hands are over the hand size and a real
// cleanup would stop on the discard.
func endTurn(t *testing.T, g *game.Game) {
	t.Helper()
	if err := g.EndTurnNowForTest(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
}

// assertTurnOf fails unless it is `seat`'s turn and its extra-ness is `extra`.
func assertTurnOf(t *testing.T, g *game.Game, seat int, extra bool, when string) {
	t.Helper()
	if g.Turn.ActiveSeat != seat || g.Turn.Extra != extra {
		t.Fatalf("%s: seat %d extra=%v, want seat %d extra=%v", when, g.Turn.ActiveSeat, g.Turn.Extra, seat, extra)
	}
}

func TestTimeWarpGivesTheTargetPlayerTheNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	opp := (me + 2) % len(g.Seats)
	castCatalogSpell(t, g, "Time Warp", "Sorcery", timeWarpOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[opp].ID}})
	passPriorityAroundTable(t, g)

	endTurn(t, g)
	assertTurnOf(t, g, opp, true, "after Time Warp's turn")
	endTurn(t, g)
	assertTurnOf(t, g, (me+1)%len(g.Seats), false, "after the extra turn (rotation resumes after the caster)")
}

func TestTimeStretchGivesTwoTurns(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	castCatalogSpell(t, g, "Time Stretch", "Sorcery", timeStretchOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[me].ID}})
	passPriorityAroundTable(t, g)

	endTurn(t, g)
	assertTurnOf(t, g, me, true, "first Time Stretch turn")
	endTurn(t, g)
	assertTurnOf(t, g, me, true, "second Time Stretch turn")
	endTurn(t, g)
	assertTurnOf(t, g, (me+1)%len(g.Seats), false, "after both extra turns")
}

func TestTakeAnExtraTurnSorceries(t *testing.T) {
	for _, c := range []struct{ name, oracle string }{
		{"Temporal Manipulation", temporalManipulationOracle},
		{"Capture of Jingzhou", captureOfJingzhouOracle},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Turn.ActiveSeat
			spell := castCatalogSpell(t, g, c.name, "Sorcery", c.oracle, nil)
			passPriorityAroundTable(t, g)
			if !g.Seats[me].Graveyard.Contains(spell) {
				t.Errorf("%s should be in its owner's graveyard", c.name)
			}
			endTurn(t, g)
			assertTurnOf(t, g, me, true, "after "+c.name)
		})
	}
}

// Temporal Mastery exiles itself, so it is not left in the graveyard.
func TestTemporalMasteryTakesATurnAndExilesItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	spell := castCatalogSpell(t, g, "Temporal Mastery", "Sorcery", temporalMasteryOracle, nil)
	passPriorityAroundTable(t, g)
	if g.Seats[me].Graveyard.Contains(spell) || !g.Exile.Contains(spell) {
		t.Errorf("Temporal Mastery should be in exile (graveyard %v, exile %v)",
			g.Seats[me].Graveyard.Contains(spell), g.Exile.Contains(spell))
	}
	endTurn(t, g)
	assertTurnOf(t, g, me, true, "after Temporal Mastery")
}

// Final Fortune's loss fires in the EXTRA turn's end step: not in the
// end step of the turn it was cast in, which an unbound "next end step"
// trigger would have picked, and then the player is out.
func TestFinalFortuneLosesAtTheExtraTurnsEndStep(t *testing.T) {
	for _, c := range []struct{ name, typeLine, oracle string }{
		{"Final Fortune", "Instant", finalFortuneOracle},
		{"Last Chance", "Sorcery", lastChanceOracle},
		{"Warrior's Oath", "Sorcery", warriorsOathOracle},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Turn.ActiveSeat
			castCatalogSpell(t, g, c.name, c.typeLine, c.oracle, nil)
			passPriorityAroundTable(t, g)

			advanceTo(t, g, game.StepEnd)
			passPriorityAroundTable(t, g)
			if g.Seats[me].Eliminated {
				t.Fatal("lost in the end step of the turn the spell resolved in")
			}
			if len(g.DelayedTriggers) != 1 {
				t.Fatalf("the loss trigger should still be waiting: %d queued", len(g.DelayedTriggers))
			}

			endTurn(t, g)
			assertTurnOf(t, g, me, true, "the extra turn")
			advanceTo(t, g, game.StepEnd)
			passPriorityAroundTable(t, g)
			if !g.Seats[me].Eliminated {
				t.Fatal("did not lose at the beginning of the extra turn's end step")
			}
			if g.State != game.StateActive {
				t.Fatalf("three players remain; the game should go on: %s", g.State)
			}
			assertTurnOf(t, g, (me+1)%len(g.Seats), false, "after the loser's extra turn")
		})
	}
}

// A second extra turn taken during Final Fortune's turn comes AFTER
// it (CR 500.7 only reorders turns not yet begun), so it does not
// rescue the player; the loss still fires in the first extra turn.
func TestFinalFortuneIsNotRescuedByALaterExtraTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	castCatalogSpell(t, g, "Final Fortune", "Instant", finalFortuneOracle, nil)
	passPriorityAroundTable(t, g)
	endTurn(t, g)
	castCatalogSpell(t, g, "Temporal Manipulation", "Sorcery", temporalManipulationOracle, nil)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if !g.Seats[me].Eliminated {
		t.Fatal("the later extra turn should not have delayed Final Fortune's loss")
	}
}

func TestMagistratesScepterChargesAndTakesATurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	scepter := b12Push(g, me.ID, "Magistrate's Scepter", "Artifact", magistratesScepterOracle, 0, 0)
	toMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, scepter, 1, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the extra-turn ability activated without three charge counters")
	}
	if err := g.ActivateCatalogAbility(me.ID, scepter, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("charge: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, scepter, game.CounterCharge); got != 1 {
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
	if got := counterCount(g, scepter, game.CounterCharge); got != 0 {
		t.Errorf("charge counters %d after paying the cost, want 0", got)
	}
	passPriorityAroundTable(t, g)
	endTurn(t, g)
	assertTurnOf(t, g, me.Seat, true, "after the Scepter")
}

func TestTeferiMasterOfTimeMinusTenTakesTwoTurns(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	teferi := pushCatalogWalker(g, me.ID, "Teferi, Master of Time", teferiMasterOfTimeOracle, 10)
	toMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, teferi, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−10: %v", err)
	}
	passPriorityAroundTable(t, g)
	endTurn(t, g)
	assertTurnOf(t, g, me.Seat, true, "first −10 turn")
	endTurn(t, g)
	assertTurnOf(t, g, me.Seat, true, "second −10 turn")
	endTurn(t, g)
	assertTurnOf(t, g, (me.Seat+1)%len(g.Seats), false, "after both turns")
}

// Avatar Kuruk's exhaust ability takes a turn, and only once.
func TestAvatarKurukExhaustTakesAnExtraTurnOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kuruk := b12Push(g, me.ID, "Avatar Kuruk", "Legendary Creature — Avatar", theLegendOfKurukOracleID+"#1", 4, 3)
	toMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, kuruk, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Waterbend {20}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, kuruk, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("an exhaust ability activated twice")
	}
	endTurn(t, g)
	assertTurnOf(t, g, me.Seat, true, "after Avatar Kuruk's ability")
	endTurn(t, g)
	assertTurnOf(t, g, (me.Seat+1)%len(g.Seats), false, "after the extra turn")
}
