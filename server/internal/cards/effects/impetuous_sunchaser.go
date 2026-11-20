package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Impetuous Sunchaser — Creature — Human Soldier {1}{R}, 1/1 (Born of the Gods,
// #5559):
//
//	"Flying, haste
//	 This creature attacks each combat if able."
//
// Flamewake Phoenix's keywords and attack requirement (AttacksEachCombat).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7f240b1f-fb7b-49a3-bcaf-6348956d2a38",
		Name:            "Impetuous Sunchaser",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Static:          []game.StaticAbility{AttacksEachCombat()},
	})
}
