package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tunneling Geopede — Creature — Insect {2}{R}, 3/2 (Battle for
// Zendikar, #5547):
//
//	"Landfall — Whenever a land you control enters, this creature
//	 deals 1 damage to each opponent."
//
// Spitfire Lagac's landfall: only a land entering under its
// controller's control triggers it, and every opponent is dealt the
// damage at once as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "78cd6d20-033f-471f-a86e-59ec8307ae1f",
		Name:         "Tunneling Geopede",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			LandfallPaying("Tunneling Geopede — 1 damage to each opponent", game.LandfallPayoff{DamageEachOpponent: 1}, func(g *game.Game, item *game.StackItem) error {
				return damageToEachOpponent(g, item, 1)
			}),
		},
	})
}
