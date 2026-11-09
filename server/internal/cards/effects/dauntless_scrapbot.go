package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "b75ca527-4494-4957-b2e3-8c590df42e95",
		Name:         "Dauntless Scrapbot",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dauntless Scrapbot — exile each opponent's graveyard; create a Lander token",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, opp := range ctx.Opponents() {
						if err := exileGraveyardForEffect(g, item, opp); err != nil {
							return err
						}
					}
					return CreateToken{Template: b16LanderToken(), N: 1}.Apply(ctx)
				}),
		},
	})
}
