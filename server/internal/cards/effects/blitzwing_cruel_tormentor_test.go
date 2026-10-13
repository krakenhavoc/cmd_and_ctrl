package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func blitzwingRow() cards.Card {
	return transformerRow(blitzwingOracleID,
		disturbFace{"Blitzwing, Cruel Tormentor", "Legendary Artifact Creature — Robot", "{5}{B}", "6", "5"},
		disturbFace{"Blitzwing, Adaptive Assailant", "Legendary Artifact — Vehicle", "", "3", "5"},
		"More Than Meets the Eye {3}{B}\nAt the beginning of your end step, target opponent loses life equal to the life that player lost this turn. If no life is lost this way, convert Blitzwing.",
		"Living metal\nAt the beginning of combat on your turn, choose flying or indestructible at random. Blitzwing gains that ability until end of turn.\n"+
			"Whenever Blitzwing deals combat damage to a player, convert it.",
		"Living metal", "Convert")
}

// The end-step trigger drains the life the target already lost this
// turn, and does not convert.
func TestBlitzwingDoublesTheLifeAnOpponentLostAndStaysPut(t *testing.T) {
	g, me, opp := flipMain(t)
	id := transformerIn(t, g, blitzwingRow(), me)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(id, opp.ID, -3); err != nil {
			t.Fatal(err)
		}
	})
	life := opp.Life
	advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	b04WaitForPick(t, g, me.ID)
	b16PickPlayer(t, g, me.ID, opp.ID)
	tfSettle(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d → %d, want another 3 lost", life, opp.Life)
	}
	wantFace(t, g, id, 0, "Blitzwing, Cruel Tormentor")
}

// "If no life is lost this way, convert Blitzwing": the target lost
// nothing this turn.
func TestBlitzwingConvertsWhenNoLifeIsLost(t *testing.T) {
	g, me, opp := flipMain(t)
	id := transformerIn(t, g, blitzwingRow(), me)
	life := opp.Life
	advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	b04WaitForPick(t, g, me.ID)
	b16PickPlayer(t, g, me.ID, opp.ID)
	tfSettle(t, g)
	if opp.Life != life {
		t.Errorf("life %d → %d, want unchanged", life, opp.Life)
	}
	wantFace(t, g, id, 1, "Blitzwing, Adaptive Assailant")
}

// Only the TARGET's loss counts: another opponent's loss this turn does
// not keep Blitzwing from converting.
func TestBlitzwingReadsOnlyTheTargetsLoss(t *testing.T) {
	g, me, opp := flipMain(t)
	id := transformerIn(t, g, blitzwingRow(), me)
	other := g.Seats[2]
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(id, other.ID, -5); err != nil {
			t.Fatal(err)
		}
	})
	advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	b04WaitForPick(t, g, me.ID)
	b16PickPlayer(t, g, me.ID, opp.ID)
	tfSettle(t, g)
	wantFace(t, g, id, 1, "Blitzwing, Adaptive Assailant")
}

// At the beginning of combat on its controller's turn the Vehicle gains
// exactly one of flying and indestructible.
func TestBlitzwingAdaptiveAssailantGainsOneAbilityAtRandom(t *testing.T) {
	g, me, _ := flipMain(t)
	id := transformerConverted(t, g, blitzwingRow(), me)
	advanceTo(t, g, game.StepBeginCombat)
	tfSettle(t, g)
	c := flipState(t, g, id)
	fly, ind := game.HasKeyword(&c, "flying"), game.HasKeyword(&c, "indestructible")
	if fly == ind {
		t.Fatalf("flying=%v indestructible=%v, want exactly one", fly, ind)
	}
	if c.ActiveFace != 1 {
		t.Errorf("the combat trigger turned it over (face %d)", c.ActiveFace)
	}
}

// Combat damage to a player converts the Vehicle back, once.
func TestBlitzwingAdaptiveAssailantConvertsOnCombatDamage(t *testing.T) {
	g, me, opp := flipMain(t)
	id := transformerConverted(t, g, blitzwingRow(), me)
	// Past the beginning of combat, so its own trigger is not queued
	// beside the one under test.
	advanceTo(t, g, game.StepBeginCombat)
	tfSettle(t, g)
	emitCombatDamage(g, id, opp.ID, 3)
	tfSettle(t, g)
	c := wantFace(t, g, id, 0, "Blitzwing, Cruel Tormentor")
	if c.TurnOvers != 2 {
		t.Errorf("turn-overs = %d, want 2", c.TurnOvers)
	}
}
