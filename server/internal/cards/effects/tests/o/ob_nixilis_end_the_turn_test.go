package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #2385 — Ob Nixilis, Captive Kingpin's "until your next end step" window
// stays open when the turn is ended (Sundial of the Infinite, CR 724.1)
// before that end step, and closes at the player's next end step that
// begins (CR 611.2b, CR 724.1d).
func TestObNixilisWindowSurvivesSundialEndingTheTurn(t *testing.T) {
	g, me, ob, top := effectstest.ObKingpinBoard(t)
	sundial := effectstest.PushCatalogPermanent(g, me.ID, "Sundial of the Infinite", "Artifact", effectstest.SundialOracle, false)
	g.WithWriteLock(func() {
		for _, opp := range g.Seats[1:3] {
			_ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1)
		}
	})
	effectstest.SettleLifeBoundary(g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) || effectstest.CountersOn(g, ob, "+1/+1") != 1 {
		t.Fatal("trigger did not resolve")
	}
	if !effectstest.WindowAt(g, top, me.ID) {
		t.Fatal("the window is closed on the turn it was made")
	}

	turn := g.Turn.Seq
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, sundial, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Sundial: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DiscardIfOwed(t, g)
	if g.Turn.Seq == turn || g.Turn.ActiveSeat == me.Seat {
		t.Fatalf("the turn did not end: turn %d seat %d step %s", g.Turn.Seq, g.Turn.ActiveSeat, g.Turn.Step)
	}
	if !effectstest.WindowAt(g, top, me.ID) {
		t.Fatal("the ended turn's skipped end step closed the window")
	}

	for seat := 1; seat < len(g.Seats); seat++ {
		effectstest.AdvanceToStepOf(t, g, seat, game.StepEnd)
		if !effectstest.WindowAt(g, top, me.ID) {
			t.Fatalf("seat %d's end step closed my window", seat)
		}
	}
	effectstest.AdvanceToStepOf(t, g, me.Seat, game.StepPostcombatMain)
	if !effectstest.WindowAt(g, top, me.ID) {
		t.Fatal("the window closed before my next end step")
	}
	effectstest.AdvanceToStepOf(t, g, me.Seat, game.StepEnd)
	if effectstest.WindowAt(g, top, me.ID) {
		t.Error("the window outlived my next end step")
	}
}
