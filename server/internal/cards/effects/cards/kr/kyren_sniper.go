package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kyren Sniper — Creature — Goblin {2}{R}, 1/1 (Mercadian Masques, S82 OLD,
// #5619):
//
//	"At the beginning of your upkeep, you may have this creature deal 1 damage
//	 to target player or planeswalker."
//
// The target is chosen as the trigger goes on the stack and "you may" is asked
// as it resolves, so a declined trigger deals nothing. The damage is the
// Sniper's, read from its last-known information if it has left (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7661dda6-4e7b-42d0-8374-0a1fe65b83e4",
		Name:         "Kyren Sniper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.AtYourUpkeep("Kyren Sniper — 1 damage to target player or planeswalker",
				effects.SourceDealsDamageToEachLegalTarget(1)), effects.TargetPlayerOrPlaneswalker()),
				"Kyren Sniper — deal 1 damage to target player or planeswalker?"),
		},
	})
}
