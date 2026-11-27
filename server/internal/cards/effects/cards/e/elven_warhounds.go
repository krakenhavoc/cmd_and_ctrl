package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elven Warhounds — Creature — Dog {3}{G}, 2/2 (Tempest, S82 OLD, #5628):
//
//	"Whenever this creature becomes blocked by a creature, put that creature on
//	 top of its owner's library."
//
// Somberwald Vigilante's per-blocker trigger (CR 509.3d: one for each blocking
// creature). As it resolves, that blocker goes to the top of its owner's
// library only if it is still on the battlefield as the object it was (CR 400.7);
// the Hounds stay blocked either way (CR 509.1h), so nothing else blocks.
//
// No simplification.
func init() {
	const name = "Elven Warhounds"
	effects.Register(effects.Spec{
		OracleID:     "ae997fee-2c1f-43c8-9dba-8b409d8e449e",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, effects.ViciousBattleragerBlocked, name+" — put that creature on top of its owner's library",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || item.Trigger.Object == nil {
						return nil
					}
					o := item.Trigger.Object
					if _, ok := g.PermanentForEffect(game.ObjectRef{ID: o.ID, Epoch: o.Epoch}); !ok {
						return nil
					}
					return g.TuckToLibraryForEffect(o.ID, false)
				}),
		},
	})
}
