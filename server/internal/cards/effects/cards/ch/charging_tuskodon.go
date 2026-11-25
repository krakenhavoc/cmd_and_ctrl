package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Charging Tuskodon — Creature — Dinosaur {3}{R}{R}, 4/4 (Rivals of Ixalan,
// S82 OLD batch 217, #5531):
//
//	"Trample
//	 If this creature would deal combat damage to a player, it deals double that
//	 damage to that player instead."
//
// A replacement on its own combat damage to a player (CR 614.1): the amount is
// doubled before it is dealt, so lifelink, a damage trigger and a prevention
// see the doubled figure, and trample's excess to the player is doubled too.
// Damage to a creature or planeswalker, and noncombat damage, is untouched.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1198d47a-4235-4678-bd2e-20f2a97b5924",
		Name:            "Charging Tuskodon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements: []game.ReplacementEffect{
			effects.DoubleCombatDamageThisDealsToPlayers("Charging Tuskodon: deals double combat damage to a player"),
		},
	})
}
