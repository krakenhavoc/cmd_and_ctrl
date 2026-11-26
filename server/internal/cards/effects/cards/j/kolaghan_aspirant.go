package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kolaghan Aspirant — Creature — Human Warrior {1}{R}, 2/1 (Dragons of Tarkir):
//
//	"Whenever this creature becomes blocked by a creature, this creature deals 1
//	 damage to that creature."
//
// One trigger per blocker (CR 509.3d: it triggers for each creature that blocks
// it), and the damage goes to that blocker only if it is still on the battlefield
// as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "334178f6-cf4b-4ea6-a24f-c7375d1b38de",
		Name:         "Kolaghan Aspirant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, effects.ViciousBattleragerBlocked,
				"Kolaghan Aspirant — it deals 1 damage to the creature blocking it",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || item.Trigger.Object == nil {
						return nil
					}
					o := item.Trigger.Object
					if _, ok := g.PermanentForEffect(game.ObjectRef{ID: o.ID, Epoch: o.Epoch}); !ok {
						return nil
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: o.ID, Amount: 1}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
