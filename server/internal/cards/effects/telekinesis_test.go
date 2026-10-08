package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const telekinesisOracle = "1b9dd2b6-d14d-4c1e-9885-00ab2c0bf8da"

// #2029: Telekinesis taps the creature, prevents its combat damage this
// turn, and keeps it tapped through its controller's next TWO untap
// steps, untapping at the third.
func TestTelekinesisTapsShieldsAndFreezesTwoUntapSteps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	att := pr7Creature(g, me.ID, "Attacker", 3, "R")
	pr7bAttack(t, g, opp.ID, att)
	pr7bCastNow(t, g, "Telekinesis", "Instant", telekinesisOracle, game.CastSpellParams{Targets: pr7bTargets(att)})
	life := opp.Life
	pr7bDamageStep(t, g)
	if opp.Life != life {
		t.Fatalf("defender life %d, want %d: its combat damage is prevented", opp.Life, life)
	}
	if !mustBattlefieldCard(t, g, att).Tapped {
		t.Fatal("tapped")
	}
	for i, wantTapped := range []bool{true, true, false} {
		advanceToUpkeepOf(t, g, 1)
		advanceToUpkeepOf(t, g, 0)
		if got := mustBattlefieldCard(t, g, att).Tapped; got != wantTapped {
			t.Fatalf("after its controller's untap step %d: tapped = %v, want %v", i+1, got, wantTapped)
		}
	}
}
