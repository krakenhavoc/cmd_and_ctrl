package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Primal Empathy — Enchantment {1}{G}{U}:
//
//	"At the beginning of your upkeep, draw a card if you control a
//	 creature with the greatest power among creatures on the battlefield.
//	 Otherwise, put a +1/+1 counter on a creature you control."
//
// One upkeep trigger with an "otherwise": the condition is read as the
// ability resolves (it is not an intervening "if"), through the layers,
// and a tie for greatest power counts (CR 608.2c). The otherwise branch
// chooses a creature as it resolves and does not target; with no
// creature of yours it does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "24645d9e-5e6d-411d-be41-c2bd1659937f",
		Name:         "Primal Empathy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(AtYourUpkeep("Primal Empathy — draw a card if you control a creature with the greatest power, otherwise put a +1/+1 counter on a creature you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if rprB22ControlsGreatestPower(g, item.Controller) {
						return DrawCards{N: 1}.Apply(ctx)
					}
					return rprB22PlusOneOnAChosenCreature(ctx)
				}), game.Purpose{Draws: 1}),
		},
	})
}
