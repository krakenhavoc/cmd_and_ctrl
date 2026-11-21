package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Accursed Centaur — Creature — Zombie Centaur {B}, 2/2 (Onslaught):
//
//	"When this creature enters, sacrifice a creature."
//
// The controller chooses a creature they control as the trigger
// resolves, and the Centaur itself is one of the choices: with nothing
// else on the battlefield it sacrifices itself. The sacrifice is not
// targeted, so hexproof and protection do not keep a creature out of it.
//
// No simplification.
func init() {
	const label = "Accursed Centaur — sacrifice a creature"
	effects.Register(effects.Spec{
		OracleID:     "b1e6921e-e460-4e58-82b0-d98ee31c279d",
		Name:         "Accursed Centaur",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				g.PlayerSacrificesForEffect(item.SourceCardID, item.Controller, effects.SacrificeSpec("a creature", effects.Creature()), label)
				return nil
			}),
		},
	})
}
