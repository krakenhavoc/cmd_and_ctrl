package testst

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const telekinesisOracle = "1b9dd2b6-d14d-4c1e-9885-00ab2c0bf8da"

// #2029: Telekinesis taps the creature, prevents its combat damage this
// turn, and keeps it tapped through its controller's next TWO untap
// steps, untapping at the third.
func TestTelekinesisTapsShieldsAndFreezesTwoUntapSteps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	att := effectstest.Pr7Creature(g, me.ID, "Attacker", 3, "R")
	effectstest.Pr7bAttack(t, g, opp.ID, att)
	effectstest.Pr7bCastNow(t, g, "Telekinesis", "Instant", telekinesisOracle, game.CastSpellParams{Targets: effectstest.Pr7bTargets(att)})
	life := opp.Life
	effectstest.Pr7bDamageStep(t, g)
	if opp.Life != life {
		t.Fatalf("defender life %d, want %d: its combat damage is prevented", opp.Life, life)
	}
	if !effectstest.MustBattlefieldCard(t, g, att).Tapped {
		t.Fatal("tapped")
	}
	for i, wantTapped := range []bool{true, true, false} {
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		effectstest.AdvanceToUpkeepOf(t, g, 0)
		if got := effectstest.MustBattlefieldCard(t, g, att).Tapped; got != wantTapped {
			t.Fatalf("after its controller's untap step %d: tapped = %v, want %v", i+1, got, wantTapped)
		}
	}
}
