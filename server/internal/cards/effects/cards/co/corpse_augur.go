package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Corpse Augur — Creature — Zombie Wizard {3}{B}, 4/2:
//
//	"When this creature dies, you draw X cards and you lose X life, where X
//	 is the number of creature cards in target player's graveyard."
//
// A dies trigger that targets a player when it goes on the stack (CR 603.3d);
// X is counted once as it resolves (the Augur itself is in its owner's
// graveyard by then), the draw and the life loss both use that one X, and a
// target that has left the game makes the whole thing do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "06eb43af-637a-4f73-8071-a7036d4c8c3d",
		Name:         "Corpse Augur",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisDies("Corpse Augur — you draw X cards and you lose X life, where X is the number of creature cards in target player's graveyard",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind != game.TargetPlayer {
								continue
							}
							x := len(effects.GraveyardIDs(g, t.ID, game.Card.IsCreature))
							if x <= 0 {
								return nil
							}
							if err := (effects.DrawCards{N: x}).Apply(ctx); err != nil {
								return err
							}
							return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -x)
						}
						return nil
					}),
				effects.TargetPlayer("target player")),
		},
	})
}
