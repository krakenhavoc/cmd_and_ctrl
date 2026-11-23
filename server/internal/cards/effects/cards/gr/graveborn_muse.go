package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Graveborn Muse — Creature — Zombie Spirit {2}{B}{B}, 3/3 (Foundations
// Commander):
//
//	"At the beginning of your upkeep, you draw X cards and lose X life,
//	 where X is the number of Zombies you control."
//
// X is counted once, as the trigger resolves, and read for both halves
// (the Muse itself is a Zombie). The draw comes first, then the life loss
// (CR 608.2c). With no Zombies (the Muse left in response) X is 0 and
// nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2a65ff4-ab82-42f7-9d90-feb29132121e",
		Name:         "Graveborn Muse",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Graveborn Muse — draw X cards and lose X life, X the Zombies you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					x := effects.CmdrB3ZombiesYouControl(ctx)
					if x <= 0 {
						return nil
					}
					if err := (effects.DrawCards{N: x}).Apply(ctx); err != nil {
						return err
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, ctx.Controller(), -x)
				}),
		},
	})
}
