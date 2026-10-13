package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func cunningBanditRow() cards.Card {
	return flipRow(cunningBanditOracleID, "{1}{R}{R}", []string{"R"}, nil,
		cards.CardFace{Name: "Cunning Bandit", TypeLine: "Creature — Human Warrior",
			OracleText: "Whenever you cast a Spirit or Arcane spell, you may put a ki counter on this creature.\nAt the beginning of the end step, if there are two or more ki counters on this creature, you may flip it.",
			Power:      "2", Toughness: "2"},
		cards.CardFace{Name: "Azamuki, Treachery Incarnate", TypeLine: "Legendary Creature — Spirit",
			OracleText: "Remove a ki counter from Azamuki: Gain control of target creature until end of turn.",
			Power:      "5", Toughness: "2"})
}

func TestCunningBanditKiCountersAndFlip(t *testing.T) {
	kiCycleCheck(t, cunningBanditRow, cunningBanditOracleID, "Cunning Bandit", "Azamuki, Treachery Incarnate", []string{"R"}, 3)
}

// Azamuki steals a creature until end of turn for a ki counter, and the
// creature goes home at cleanup.
func TestAzamukiStealsUntilEndOfTurn(t *testing.T) {
	g, me, opp := flipMain(t)
	id := flipEntered(t, g, cunningBanditRow(), me)
	theirs := apaPush(g, opp.ID, opp.ID, apaCreature("Their Bear", "oracle-their-bear", 2, 2))
	pfAddCounters(g, id, "ki", 2)
	g.WithWriteLock(func() {
		if err := g.FlipPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})

	p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Targets: cardRefs(theirs)})
	if n := kiCount(t, g, id); n != 1 {
		t.Errorf("%d ki counters, want 1", n)
	}
	if c := flipState(t, g, theirs); c.Controller != me.ID {
		t.Fatalf("Their Bear is controlled by %v, want me", c.Controller)
	}
	advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	for i := 0; i < 4 && flipState(t, g, theirs).Controller == me.ID; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if c := flipState(t, g, theirs); c.Controller != opp.ID {
		t.Errorf("Their Bear is controlled by %v after cleanup, want its owner", c.Controller)
	}
}
