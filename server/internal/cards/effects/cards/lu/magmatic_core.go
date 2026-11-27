package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Magmatic Core — Enchantment {2}{R}{R} (Coldsnap, S82 OLD b558):
//
//	"Cumulative upkeep {1}
//	 At the beginning of your end step, this enchantment deals X damage divided
//	 as you choose among any number of target creatures, where X is the number
//	 of age counters on it."
//
// X is the Core's age counters as the trigger is put on the stack, so the
// division is announced for that many (CR 601.2d); a target that has gone takes
// nothing and its share is not redistributed (CR 608.2b).
//
// No simplification.
func init() {
	end := effects.AtYourEndStep("Magmatic Core — X damage divided as you choose among any number of target creatures, where X is the number of age counters on it",
		func(g *game.Game, item *game.StackItem) error {
			return effects.DealDividedDamage(effects.NewContext(g, item))
		})
	end.Targets = effects.TargetCreature("any number of target creatures").WithCount(0, 0).Dividing(effects.DivideBy(effects.DivideAgeCountersOnSource))
	effects.Register(effects.Spec{
		OracleID:     "b9526bd5-ed83-466f-8819-e9a8487bbe77",
		Name:         "Magmatic Core",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Magmatic Core — cumulative upkeep {1}", "{1}"),
			end,
		},
	})
}
