package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "583102f8-7232-4fbb-a715-087448536bb2",
		Name:             "Drix Fatemaker",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Warp("{1}{G}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Drix Fatemaker — put a +1/+1 counter on target creature", PlusOneCounterOnChosenTarget),
				TargetCreature("target creature")),
		},
		Static: []game.StaticAbility{
			CreaturesYouControlWithCountersHave(game.CounterPlusOne, false, "trample"),
		},
	})
}
