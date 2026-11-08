package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spitfire Lagac — Creature — Lizard {3}{R}, 3/4 (Foundations, #4465):
//
//	"Landfall — Whenever a land you control enters, this creature deals
//	 1 damage to each opponent."
//
// Sabotender's landfall ping without the reach: the Lagac is the damage
// source, and every opponent is dealt 1 for each land that enters under
// its controller's control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7215ca87-bbff-4c14-a1f5-de5ddd97c875",
		Name:         "Spitfire Lagac",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			LandfallPaying("Spitfire Lagac — 1 damage to each opponent", game.LandfallPayoff{DamageEachOpponent: 1}, func(g *game.Game, item *game.StackItem) error {
				return damageToEachOpponent(g, item, 1)
			}),
		},
	})
}
