package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flameborn Hellion — Creature — Hellion {5}{R}, 5/4 (Scars of Mirrodin, #5577):
//
//	"Haste
//	 This creature attacks each combat if able."
//
// Printed haste and Bloodrock Cyclops' attack requirement
// (AttacksEachCombat, CR 508.1d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "47d4da1e-b1a2-4a23-9b96-c85d5d20e3ef",
		Name:            "Flameborn Hellion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static:          []game.StaticAbility{effects.AttacksEachCombat()},
	})
}
