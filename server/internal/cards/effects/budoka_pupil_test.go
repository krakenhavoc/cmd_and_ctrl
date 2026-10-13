package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func budokaPupilRow() cards.Card {
	return flipRow(budokaPupilOracleID, "{1}{G}{G}", []string{"G"}, []string{"Trample"},
		cards.CardFace{Name: "Budoka Pupil", TypeLine: "Creature — Human Monk",
			OracleText: "Whenever you cast a Spirit or Arcane spell, you may put a ki counter on this creature.\nAt the beginning of the end step, if there are two or more ki counters on this creature, you may flip it.",
			Power:      "2", Toughness: "2"},
		cards.CardFace{Name: "Ichiga, Who Topples Oaks", TypeLine: "Legendary Creature — Spirit",
			OracleText: "Trample\nRemove a ki counter from Ichiga: Target creature gets +2/+2 until end of turn.",
			Power:      "4", Toughness: "3"})
}

// The ki cycle: counters, the "you may" flip at the end step.
func TestBudokaPupilKiCountersAndFlip(t *testing.T) {
	kiCycleCheck(t, budokaPupilRow, budokaPupilOracleID, "Budoka Pupil", "Ichiga, Who Topples Oaks", []string{"G"}, 3)
}

// Ichiga has trample as its own half's keyword, and spends a ki counter
// (the ones the Pupil collected) to give a creature +2/+2.
func TestIchigaPumpsForAKiCounter(t *testing.T) {
	g, me, opp := flipMain(t)
	id := flipEntered(t, g, budokaPupilRow(), me)
	bear := apaPush(g, opp.ID, opp.ID, apaCreature("Their Bear", "oracle-their-bear", 2, 2))
	top := flipState(t, g, id)
	if game.HasKeyword(&top, "trample") {
		t.Fatal("Budoka Pupil has trample")
	}
	pfAddCounters(g, id, "ki", 2)
	g.WithWriteLock(func() {
		if err := g.FlipPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	c := wantFlipped(t, g, id, "Ichiga, Who Topples Oaks", []string{"G"}, 3)
	if !game.HasKeyword(&c, "trample") {
		t.Errorf("Ichiga keywords %v, want trample", c.Keywords)
	}

	p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	if n := kiCount(t, g, id); n != 1 {
		t.Errorf("%d ki counters left, want 1 (one removed as the cost)", n)
	}
	if e := flipState(t, g, bear).Effective(); e.Power != 4 || e.Toughness != 4 {
		t.Errorf("Their Bear is %d/%d, want 4/4", e.Power, e.Toughness)
	}
	// With no counters left the ability cannot be activated.
	pfAddCounters(g, id, "ki", -1)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err == nil {
		t.Error("activated Ichiga with no ki counters")
	}
}
