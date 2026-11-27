package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sylvan Basilisk — {3}{G}{G} Creature — Basilisk 2/4 (Starter 1999, S82 OLD batch 702, #5620):
//
//	"Whenever this creature becomes blocked by a creature, destroy that
//	 creature."
//
// One trigger per blocker (CR 509.3d), as Kolaghan Aspirant's; the blocker is
// destroyed only if it is still the same object on the battlefield as the
// ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4f348e7a-fe13-4621-bd88-5d2fb5b597f2",
		Name:         "Sylvan Basilisk",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, effects.ViciousBattleragerBlocked,
				"Sylvan Basilisk — destroy that creature",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || item.Trigger.Object == nil {
						return nil
					}
					o := item.Trigger.Object
					if _, ok := g.PermanentForEffect(game.ObjectRef{ID: o.ID, Epoch: o.Epoch}); !ok {
						return nil
					}
					return effects.DestroyTarget{Target: o.ID}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
