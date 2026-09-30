package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const wanderingArchaicOracle = "6556c4c0-b10d-4208-821b-0c0a49abd188"

// TestWanderingArchaicDeclinePaymentThenCopy — the opponent declines
// to pay {2}; Wandering Archaic's controller may then copy the spell,
// which deals its damage a second time.
func TestWanderingArchaicDeclinePaymentThenCopy(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Wandering Archaic", "Creature — Avatar", wanderingArchaicOracle, false)

	for g.Turn.ActiveSeat != 1 || (g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain) {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}

	before := me.Life
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatalf("no pay-unless prompt for the caster")
	}
	answerPayUnless(t, g, opp.ID, false)
	answerMayChoice(t, g, me.ID, true)
	settleThroughCopyRetarget(t, g, me.ID, me.ID)

	if got := before - me.Life; got != 6 {
		t.Errorf("declined payment + copy: took %d damage, want 6 (bolt + copy)", got)
	}
}

// TestWanderingArchaicPaymentSkipsTheCopy — paying {2} avoids the
// copy question entirely.
func TestWanderingArchaicPaymentSkipsTheCopy(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Wandering Archaic", "Creature — Avatar", wanderingArchaicOracle, false)

	for g.Turn.ActiveSeat != 1 || (g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain) {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	floatMana(t, g, opp, "{C}{C}")

	before := me.Life
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, opp.ID, true)
	passPriorityAroundTable(t, g)

	if got := before - me.Life; got != 3 {
		t.Errorf("paid {2}: took %d damage, want 3 (bolt only, no copy)", got)
	}
}
