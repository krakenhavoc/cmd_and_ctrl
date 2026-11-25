package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ashmouth Hound — Creature — Elemental Dog {1}{R}, 2/1 (Jumpstart, #5510):
//
//	"Whenever this creature blocks or becomes blocked by a creature, this
//	 creature deals 1 damage to that creature."
//
// Ornery Goblin's trigger: one block event per blocker–attacker pair (CR
// 509.3b, 509.3d), so it fires once for each creature the Hound blocks or
// is blocked by. The Hound deals the damage as an object, so one that has
// left the battlefield still deals it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d40fd63b-f55d-49f6-b4e5-a8d00f22e8e1",
		Name:         "Ashmouth Hound",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.SkewerSlingerFoe(ev, source.InstanceID) != uuid.Nil
			}, "Ashmouth Hound — 1 damage to that creature", effects.SkewerSlingerPokes),
		},
	})
}
