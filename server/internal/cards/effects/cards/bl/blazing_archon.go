package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blazing Archon — Creature — Archon {6}{W}{W}{W}, 5/6 (Ravnica: City of
// Guilds, #5600):
//
//	"Flying
//	 Creatures can't attack you."
//
// CreaturesMatchingCantAttackYou with no narrowing query, so it binds
// every creature: "you" is the player alone, so creatures may still
// attack your planeswalkers and battles, and "you" follows the Archon's
// controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "82a70e4c-01ed-4280-80e4-86cb2bf6e63d",
		Name:            "Blazing Archon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.CreaturesMatchingCantAttackYou("Creatures can't attack you.")},
	})
}
