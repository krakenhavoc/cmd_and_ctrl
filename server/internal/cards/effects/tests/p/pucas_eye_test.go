package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const pucasEyeOracle = "499b4cf9-2aaa-45a0-8362-77b727b7f7cc"

func TestPucasEyeIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(pucasEyeOracle)
	if !ok || spec.Name != "Puca's Eye" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// pucasEyeEnter casts the Eye, checks the draw, and names `color`.
func pucasEyeEnter(t *testing.T, g *game.Game, me *game.Player, color string) uuid.UUID {
	t.Helper()
	effectstest.ToMainForCost(t, g)
	hand := me.Hand.Size()
	eye := effectstest.RfdCast(t, g, "Puca's Eye", "Artifact", pucasEyeOracle, 0, 0)
	effectstest.MonSettle(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("hand %d after casting the Eye, want %d (one drawn)", got, hand+1)
	}
	effectstest.AnswerColor(t, g, me.ID, color)
	effectstest.MonSettle(t, g)
	return eye
}

// The draw comes first, then the Eye becomes the chosen colour and
// stays it past the turn.
func TestPucasEyeDrawsAndBecomesTheChosenColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	eye := pucasEyeEnter(t, g, me, "G")
	if got := effectstest.TmtB9Colors(t, g, eye); got != "G" {
		t.Fatalf("colors = %q, want green", got)
	}
	effectstest.AdvanceToStepOf(t, g, (seat+1)%len(g.Seats), game.StepUpkeep)
	if got := effectstest.TmtB9Colors(t, g, eye); got != "G" {
		t.Errorf("next turn: colors = %q, want still green", got)
	}
}

// The draw ability needs five colours among your permanents; the Eye's
// own chosen colour counts.
func TestPucasEyeDrawsOnlyWithFiveColors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	eye := pucasEyeEnter(t, g, me, "G")
	effectstest.B27Push(g, me.ID, "White Thing", "Enchantment", "", "", 0, 0, "W")
	effectstest.B27Push(g, me.ID, "Blue Thing", "Enchantment", "", "", 0, 0, "U")
	effectstest.B27Push(g, me.ID, "Black Thing", "Enchantment", "", "", 0, 0, "B")

	effectstest.B06AddMana(me, "C", "C", "C")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, eye, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with four colors among your permanents")
	}
	if me.Hand.Size() != hand || effectstest.B16Tapped(t, g, eye) {
		t.Fatal("a refused activation drew or tapped the Eye")
	}

	effectstest.B27Push(g, me.ID, "Red Thing", "Enchantment", "", "", 0, 0, "R")
	effectstest.B16Activate(t, g, me.ID, eye, 0, game.ActivateAbilityParams{Strict: true})
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d after the draw", got, hand+1)
	}
}
