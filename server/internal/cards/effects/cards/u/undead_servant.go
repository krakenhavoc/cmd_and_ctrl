package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Undead Servant — Creature — Zombie {3}{B}, 3/2 (Core Set 2020):
//
//	"When this creature enters, create a 2/2 black Zombie creature token for
//	 each card named Undead Servant in your graveyard."
//
// The count is read as the trigger resolves, in its controller's graveyard
// only (the Servant itself is on the battlefield, not counted); with none, no
// token is made.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5bf9f397-0216-4ec9-a57b-406758dcc233",
		Name:         "Undead Servant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Undead Servant — create a 2/2 black Zombie token for each card named Undead Servant in your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					named := effects.GraveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool { return c.Name == "Undead Servant" })
					if len(named) == 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie"), N: len(named)}.Apply(ctx)
				}),
		},
	})
}
