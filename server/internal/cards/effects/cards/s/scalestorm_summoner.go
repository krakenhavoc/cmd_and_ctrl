package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scalestorm Summoner — Creature — Human Warlock {2}{R}, 3/3 (Outlaws of
// Thunder Junction, #4468):
//
//	"Whenever this creature attacks, create a 3/1 red Dinosaur creature
//	 token if you control a creature with power 4 or greater."
//
// The condition is not an intervening "if" (it follows the effect), so
// the trigger always goes on the stack and the power is read as it
// resolves, with the layers applied (b24ControlsCreatureWithPowerAtLeast,
// the Ferocious reading). The token is Poetic Ingenuity's 3/1 red
// Dinosaur; it enters neither tapped nor attacking.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "55ede1d4-593d-433c-943a-4db12eeb5fa8",
		Name:         "Scalestorm Summoner",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Scalestorm Summoner — create a 3/1 red Dinosaur if you control a creature with power 4 or greater",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.B24ControlsCreatureWithPowerAtLeast(g, item.Controller, 4) {
						return nil
					}
					return effects.CreateToken{Template: effects.TokenCard("3/1 red Dinosaur"), N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
