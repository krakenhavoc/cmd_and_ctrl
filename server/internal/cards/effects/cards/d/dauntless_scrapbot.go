package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dauntless Scrapbot — Artifact Creature — Robot {3}, 3/1 (Edge of
// Eternities, #4461):
//
//	"When this creature enters, exile each opponent's graveyard. Create a
//	 Lander token."
//
// Bojuka Bog's per-graveyard exile over each of the controller's
// opponents still in the game (the controller's own graveyard stays),
// then Glacier Godmaw's Lander (b16LanderToken). Both halves are one
// trigger's resolution, in printed order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b75ca527-4494-4957-b2e3-8c590df42e95",
		Name:         "Dauntless Scrapbot",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Dauntless Scrapbot — exile each opponent's graveyard; create a Lander token",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, opp := range ctx.Opponents() {
						if err := effects.ExileGraveyardForEffect(g, item, opp); err != nil {
							return err
						}
					}
					return effects.CreateToken{Template: effects.B16LanderToken(), N: 1}.Apply(ctx)
				}),
		},
	})
}
