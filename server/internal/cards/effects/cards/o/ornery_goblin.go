package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ornery Goblin — Creature — Goblin Warrior {1}{R}, 2/1 (Ravnica: Clue
// Edition, Time Spiral Remastered):
//
//	"Whenever this creature blocks or becomes blocked by a creature, this
//	 creature deals 1 damage to that creature."
//
// Skewer Slinger's trigger without the reach: one block event per
// blocker–attacker pair (CR 509.3b, 509.3d), so it fires once for each
// creature the Goblin blocks or is blocked by. The Goblin deals the damage
// as an object, so one that has left the battlefield still deals it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cbd4e160-9eb1-42de-aee8-6f1d1398f633",
		Name:         "Ornery Goblin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.SkewerSlingerFoe(ev, source.InstanceID) != uuid.Nil
			}, "Ornery Goblin — 1 damage to that creature", effects.SkewerSlingerPokes),
		},
	})
}
