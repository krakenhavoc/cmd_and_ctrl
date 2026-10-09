package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "234ff22f-2ff1-4a73-a7c5-e9c53557c4c6",
		Name:         "Acquisition Octopus",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, ThisOrEquippedCreatureDealsCombatDamageToAPlayer,
				"Acquisition Octopus — draw a card", Do(DrawCards{N: 1})),
		},
		Activated: Reconfigure("{2}"),
	})
}
