package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shinka Gatekeeper — Creature — Ogre Warrior {2}{R}, 3/2 (Betrayers of
// Kamigawa, #5602):
//
//	"Whenever this creature is dealt damage, it deals that much damage to
//	 you."
//
// Barbed Servitor's damage trigger (SelfWasDealtDamage, ThatMuchDamage):
// the ability waits for the damage to finish and reads the total the
// Gatekeeper was dealt at the same time (CR 120.3, CR 603.2c). The
// Gatekeeper deals it to its controller as it last existed if it has left
// (CR 608.2h).
//
// No simplification.
func init() {
	const label = "Shinka Gatekeeper — it deals that much damage to you"
	effects.Register(effects.Spec{
		OracleID:     "7792d3af-5c64-4d0a-a2b3-b903d787b0ab",
		Name:         "Shinka Gatekeeper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.B35SelfWasDealtDamage(ev, source)
			},
			Key:        label,
			AtBatchEnd: effects.ThatMuchDamage(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil {
					return nil
				}
				return effects.DamageController{N: item.Trigger.Event.Amount}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
