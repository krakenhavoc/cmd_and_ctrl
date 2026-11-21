package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Acquisition Octopus — Artifact Creature — Equipment Octopus {2}{U}, 2/2:
//
//	"Whenever this creature or equipped creature deals combat damage to
//	 a player, draw a card.
//	 Reconfigure {2}"
//
// Reconfigure is reconfigure.go (#2639).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "234ff22f-2ff1-4a73-a7c5-e9c53557c4c6",
		Name:         "Acquisition Octopus",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisOrEquippedCreatureDealsCombatDamageToAPlayer,
				"Acquisition Octopus — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		Activated: effects.Reconfigure("{2}"),
	})
}
