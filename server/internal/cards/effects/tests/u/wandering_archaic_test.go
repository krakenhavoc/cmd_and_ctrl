package testsu

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestWanderingArchaicDeclinePaymentThenCopy — the opponent declines
// to pay {2}; Wandering Archaic's controller may then copy the spell,
// which deals its damage a second time.
func TestWanderingArchaicDeclinePaymentThenCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Wandering Archaic", "Creature — Avatar", effectstest.WanderingArchaicOracle, false)

	for g.Turn.ActiveSeat != 1 || (g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain) {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}

	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatalf("no pay-unless prompt for the caster")
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.SettleThroughCopyRetarget(t, g, me.ID, me.ID)

	if got := before - me.Life; got != 6 {
		t.Errorf("declined payment + copy: took %d damage, want 6 (bolt + copy)", got)
	}
}

// TestWanderingArchaicPaymentSkipsTheCopy — paying {2} avoids the
// copy question entirely.
func TestWanderingArchaicPaymentSkipsTheCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Wandering Archaic", "Creature — Avatar", effectstest.WanderingArchaicOracle, false)

	for g.Turn.ActiveSeat != 1 || (g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain) {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.FloatMana(t, g, opp, "{C}{C}")

	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, opp.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	if got := before - me.Life; got != 3 {
		t.Errorf("paid {2}: took %d damage, want 3 (bolt only, no copy)", got)
	}
}
