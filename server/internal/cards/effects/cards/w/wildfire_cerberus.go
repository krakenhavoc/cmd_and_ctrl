package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wildfire Cerberus — Creature — Dog {4}{R}, 4/3 (Journey into Nyx, S82 OLD,
// #5558):
//
//	"{5}{R}{R}: Monstrosity 1. (If this creature isn't monstrous, put a +1/+1
//	 counter on it and it becomes monstrous.)
//	 When this creature becomes monstrous, it deals 2 damage to each opponent
//	 and each creature they control."
//
// Goblin Chainwhirler's single damage instance over a becomes-monstrous
// trigger (Arbor Colossus). The creatures are those an opponent controls as
// the trigger resolves; the Cerberus is the source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2e6be9ad-d726-4598-92a7-1ba7eda317a5",
		Name:         "Wildfire Cerberus",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Monstrosity(effects.ManaCost("{5}{R}{R}"), 1)},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.WhenBecomesMonstrous("Wildfire Cerberus — 2 damage to each opponent and each creature they control",
					func(g *game.Game, item *game.StackItem) error {
						return effects.DamageEachMatchingAndEachOpponent(effects.NewContext(g, item), effects.And(effects.Creature(), effects.OpponentControls()), 2)
					}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, OpponentsOnly: true}}),
		},
	})
}
