package testsex

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	cardsex "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/ex"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func faithfulSquireRow() cards.Card {
	return effectstest.FlipRow(cardsex.FaithfulSquireOracleID, "{1}{W}{W}", []string{"W"}, []string{"Flying"},
		cards.CardFace{Name: "Faithful Squire", TypeLine: "Creature — Human Soldier",
			OracleText: "Whenever you cast a Spirit or Arcane spell, you may put a ki counter on this creature.\nAt the beginning of the end step, if there are two or more ki counters on this creature, you may flip it.",
			Power:      "2", Toughness: "2"},
		cards.CardFace{Name: "Kaiso, Memory of Loyalty", TypeLine: "Legendary Creature — Spirit",
			OracleText: "Flying\nRemove a ki counter from Kaiso: Prevent all damage that would be dealt to target creature this turn.",
			Power:      "3", Toughness: "4"})
}

func TestFaithfulSquireKiCountersAndFlip(t *testing.T) {
	effectstest.KiCycleCheck(t, faithfulSquireRow, cardsex.FaithfulSquireOracleID, "Faithful Squire", "Kaiso, Memory of Loyalty", []string{"W"}, 3)
}

// Kaiso flies (its own half's keyword; the Squire does not), and a ki
// counter prevents all damage to a creature for the turn.
func TestKaisoFliesAndShieldsACreature(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	id := effectstest.FlipEntered(t, g, faithfulSquireRow(), me)
	bear := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("My Bear", "oracle-my-bear", 2, 2))
	top := effectstest.FlipState(t, g, id)
	if game.HasKeyword(&top, "flying") {
		t.Fatal("Faithful Squire has flying")
	}
	effectstest.PfAddCounters(g, id, "ki", 2)
	g.WithWriteLock(func() {
		if err := g.FlipPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	c := effectstest.WantFlipped(t, g, id, "Kaiso, Memory of Loyalty", []string{"W"}, 3)
	if !game.HasKeyword(&c, "flying") {
		t.Errorf("Kaiso keywords %v, want flying", c.Keywords)
	}

	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	if n := effectstest.KiCount(t, g, id); n != 1 {
		t.Errorf("%d ki counters, want 1", n)
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(opp.ID, bear, 3); err != nil {
			t.Fatal(err)
		}
	})
	if b := effectstest.FlipState(t, g, bear); b.DamageMarked != 0 {
		t.Errorf("My Bear has %d damage marked, want it prevented", b.DamageMarked)
	}
}
