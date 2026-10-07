package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// bounce_land_autotap_test.go is #2461's enumerator half: the move
// list asks the same auto-tap planner the cast pays with, so a bounce
// land's two mana are two different pips here too. A cast the engine
// refuses is not offered, and every cast offered is accepted.

const (
	oracleSimicGrowthChamber = "046f5783-cc7b-416a-8cf6-2bcef9c2cc1a"
	oracleIzzetBoilerworks   = "1cb9d94a-3039-4f2e-8fcc-6996f9a45f74"
)

func TestBounceLandIsNotOfferedAsTwoOfOneColour(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	battlefieldCard(g, active, game.Card{Name: "Simic Growth Chamber", TypeLine: "Land", OracleID: oracleSimicGrowthChamber})
	uu := handCard(active, game.Card{Name: "Blue Blue", TypeLine: "Instant", ManaCost: "{U}{U}"})
	gu := handCard(active, game.Card{Name: "Green Blue", TypeLine: "Instant", ManaCost: "{G}{U}"})
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	if got := castMovesFor(moves, uu); len(got) != 0 {
		t.Errorf("{U}{U} offered off one Simic Growth Chamber: %v", labels(got))
	}
	offered := castMovesFor(moves, gu)
	if len(offered) == 0 {
		t.Error("{G}{U} not offered off a Simic Growth Chamber")
	}
	dispatchAll(t, g, active.ID, offered)
}

func TestBoilerworksAndIslandAreNotOfferedForTwoRed(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	battlefieldCard(g, active, game.Card{Name: "Izzet Boilerworks", TypeLine: "Land", OracleID: oracleIzzetBoilerworks})
	lands(g, active, "Island", "Island", 1)
	rr := handCard(active, creature("Two Red", "{1}{R}{R}", 2, 2))
	ur := handCard(active, creature("Blue Red", "{1}{U}{R}", 2, 2))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	if got := castMovesFor(moves, rr); len(got) != 0 {
		t.Errorf("{1}{R}{R} offered off Izzet Boilerworks and an Island: %v", labels(got))
	}
	offered := castMovesFor(moves, ur)
	if len(offered) == 0 {
		t.Error("{1}{U}{R} not offered off Izzet Boilerworks and an Island")
	}
	dispatchAll(t, g, active.ID, offered)
}
