package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func hiredMuscleRow() cards.Card {
	return flipRow(hiredMuscleOracleID, "{1}{B}{B}", []string{"B"}, nil,
		cards.CardFace{Name: "Hired Muscle", TypeLine: "Creature — Human Warrior",
			OracleText: "Whenever you cast a Spirit or Arcane spell, you may put a ki counter on this creature.\nAt the beginning of the end step, if there are two or more ki counters on this creature, you may flip it.",
			Power:      "2", Toughness: "2"},
		cards.CardFace{Name: "Scarmaker", TypeLine: "Legendary Creature — Spirit",
			OracleText: "Remove a ki counter from Scarmaker: Target creature gains fear until end of turn. (It can't be blocked except by artifact creatures and/or black creatures.)",
			Power:      "4", Toughness: "4"})
}

func TestHiredMuscleKiCountersAndFlip(t *testing.T) {
	kiCycleCheck(t, hiredMuscleRow, hiredMuscleOracleID, "Hired Muscle", "Scarmaker", []string{"B"}, 3)
}

// Scarmaker gives a creature fear until end of turn for a ki counter.
func TestScarmakerGrantsFear(t *testing.T) {
	g, me, _ := flipMain(t)
	id := flipEntered(t, g, hiredMuscleRow(), me)
	bear := apaPush(g, me.ID, me.ID, apaCreature("My Bear", "oracle-my-bear", 2, 2))
	pfAddCounters(g, id, "ki", 2)
	g.WithWriteLock(func() {
		if err := g.FlipPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	wantFlipped(t, g, id, "Scarmaker", []string{"B"}, 3)

	p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	if n := kiCount(t, g, id); n != 1 {
		t.Errorf("%d ki counters, want 1", n)
	}
	b := flipState(t, g, bear)
	if !game.HasKeyword(&b, "fear") {
		t.Errorf("My Bear has no fear: %+v", b.Effective().Abilities)
	}
}
