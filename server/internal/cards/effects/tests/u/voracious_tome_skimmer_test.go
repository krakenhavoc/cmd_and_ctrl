package testsu

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

const voraciousTomeSkimmerOracle = "25b0876e-c995-486e-8536-6e68b76b9c86"

func TestVoraciousTomeSkimmerIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(voraciousTomeSkimmerOracle)
	if !ok || spec.Name != "Voracious Tome-Skimmer" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// A spell cast on an opponent's turn offers the life payment: paying
// draws a card, declining does nothing.
func TestVoraciousTomeSkimmerPaysALifeToDrawOnAnOpponentsTurn(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[1]
		effectstest.B12Push(g, me.ID, "Voracious Tome-Skimmer", "Creature — Faerie Rogue", voraciousTomeSkimmerOracle, 2, 3)
		effectstest.AangAdvanceToMain(t, g, 0)
		life, hand := me.Life, me.Hand.Size()
		effectstest.KataraCastInstant(t, g, me)
		if !effectstest.HasPayUnlessFor(g, me.ID) {
			t.Fatalf("pay %v: the cast trigger did not offer the life payment", pay)
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		wantLife, wantHand := life-1, hand+1
		if !pay {
			wantLife, wantHand = life, hand
		}
		if me.Life != wantLife || me.Hand.Size() != wantHand {
			t.Errorf("pay %v: life %d -> %d, hand %d -> %d; want life %d, hand %d",
				pay, life, me.Life, hand, me.Hand.Size(), wantLife, wantHand)
		}
	}
}

// A spell cast on your own turn, or an opponent's spell, does not trigger.
func TestVoraciousTomeSkimmerIgnoresYourTurnAndOpponentsSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp, me := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Voracious Tome-Skimmer", "Creature — Faerie Rogue", voraciousTomeSkimmerOracle, 2, 3)

	effectstest.AangAdvanceToMain(t, g, 0)
	effectstest.KataraCastInstant(t, g, opp)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("an opponent's spell offered the payment")
	}

	effectstest.AangAdvanceToMain(t, g, 1)
	life := me.Life
	effectstest.KataraCastInstant(t, g, me)
	if effectstest.HasPayUnlessFor(g, me.ID) || me.Life != life {
		t.Error("a spell cast on your own turn offered the payment")
	}
}
