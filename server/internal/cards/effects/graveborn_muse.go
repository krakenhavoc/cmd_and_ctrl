package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d2a65ff4-ab82-42f7-9d90-feb29132121e",
		Name:         "Graveborn Muse",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Graveborn Muse — draw X cards and lose X life, X the Zombies you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					x := cmdrB3ZombiesYouControl(ctx)
					if x <= 0 {
						return nil
					}
					if err := (DrawCards{N: x}).Apply(ctx); err != nil {
						return err
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, ctx.Controller(), -x)
				}),
		},
	})
}
