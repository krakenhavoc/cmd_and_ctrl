package testsg

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	cardsh "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/h"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func hiredMuscleRow() cards.Card {
	return effectstest.FlipRow(cardsh.HiredMuscleOracleID, "{1}{B}{B}", []string{"B"}, nil,
		cards.CardFace{Name: "Hired Muscle", TypeLine: "Creature — Human Warrior",
			OracleText: "Whenever you cast a Spirit or Arcane spell, you may put a ki counter on this creature.\nAt the beginning of the end step, if there are two or more ki counters on this creature, you may flip it.",
			Power:      "2", Toughness: "2"},
		cards.CardFace{Name: "Scarmaker", TypeLine: "Legendary Creature — Spirit",
			OracleText: "Remove a ki counter from Scarmaker: Target creature gains fear until end of turn. (It can't be blocked except by artifact creatures and/or black creatures.)",
			Power:      "4", Toughness: "4"})
}

func TestHiredMuscleKiCountersAndFlip(t *testing.T) {
	effectstest.KiCycleCheck(t, hiredMuscleRow, cardsh.HiredMuscleOracleID, "Hired Muscle", "Scarmaker", []string{"B"}, 3)
}

// Scarmaker gives a creature fear until end of turn for a ki counter.
func TestScarmakerGrantsFear(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	id := effectstest.FlipEntered(t, g, hiredMuscleRow(), me)
	bear := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("My Bear", "oracle-my-bear", 2, 2))
	effectstest.PfAddCounters(g, id, "ki", 2)
	g.WithWriteLock(func() {
		if err := g.FlipPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.WantFlipped(t, g, id, "Scarmaker", []string{"B"}, 3)

	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	if n := effectstest.KiCount(t, g, id); n != 1 {
		t.Errorf("%d ki counters, want 1", n)
	}
	b := effectstest.FlipState(t, g, bear)
	if !game.HasKeyword(&b, "fear") {
		t.Errorf("My Bear has no fear: %+v", b.Effective().Abilities)
	}
}
