package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drix Fatemaker — Creature — Drix Wizard {3}{G}, 3/2 (Edge of
// Eternities, #4461):
//
//	"When this creature enters, put a +1/+1 counter on target creature.
//	 Each creature you control with a +1/+1 counter on it has trample.
//	 Warp {1}{G}"
//
// The enters trigger is Luminarch Aspirant's counter on any target
// creature (its own or another's, anyone's). The static is Exava's
// keyword grant over counters, "each creature you control" so it includes
// the Fatemaker once it has a counter. Warp is the shared constructor.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "583102f8-7232-4fbb-a715-087448536bb2",
		Name:             "Drix Fatemaker",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Warp("{1}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Drix Fatemaker — put a +1/+1 counter on target creature", effects.PlusOneCounterOnChosenTarget),
				effects.TargetCreature("target creature")),
		},
		Static: []game.StaticAbility{
			effects.CreaturesYouControlWithCountersHave(game.CounterPlusOne, false, "trample"),
		},
	})
}
