package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func faithfulSquireRow() cards.Card {
	return flipRow(faithfulSquireOracleID, "{1}{W}{W}", []string{"W"}, []string{"Flying"},
		cards.CardFace{Name: "Faithful Squire", TypeLine: "Creature — Human Soldier",
			OracleText: "Whenever you cast a Spirit or Arcane spell, you may put a ki counter on this creature.\nAt the beginning of the end step, if there are two or more ki counters on this creature, you may flip it.",
			Power:      "2", Toughness: "2"},
		cards.CardFace{Name: "Kaiso, Memory of Loyalty", TypeLine: "Legendary Creature — Spirit",
			OracleText: "Flying\nRemove a ki counter from Kaiso: Prevent all damage that would be dealt to target creature this turn.",
			Power:      "3", Toughness: "4"})
}

func TestFaithfulSquireKiCountersAndFlip(t *testing.T) {
	kiCycleCheck(t, faithfulSquireRow, faithfulSquireOracleID, "Faithful Squire", "Kaiso, Memory of Loyalty", []string{"W"}, 3)
}

// Kaiso flies (its own half's keyword; the Squire does not), and a ki
// counter prevents all damage to a creature for the turn.
func TestKaisoFliesAndShieldsACreature(t *testing.T) {
	g, me, opp := flipMain(t)
	id := flipEntered(t, g, faithfulSquireRow(), me)
	bear := apaPush(g, me.ID, me.ID, apaCreature("My Bear", "oracle-my-bear", 2, 2))
	top := flipState(t, g, id)
	if game.HasKeyword(&top, "flying") {
		t.Fatal("Faithful Squire has flying")
	}
	pfAddCounters(g, id, "ki", 2)
	g.WithWriteLock(func() {
		if err := g.FlipPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	c := wantFlipped(t, g, id, "Kaiso, Memory of Loyalty", []string{"W"}, 3)
	if !game.HasKeyword(&c, "flying") {
		t.Errorf("Kaiso keywords %v, want flying", c.Keywords)
	}

	p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	if n := kiCount(t, g, id); n != 1 {
		t.Errorf("%d ki counters, want 1", n)
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(opp.ID, bear, 3); err != nil {
			t.Fatal(err)
		}
	})
	if b := flipState(t, g, bear); b.DamageMarked != 0 {
		t.Errorf("My Bear has %d damage marked, want it prevented", b.DamageMarked)
	}
}
