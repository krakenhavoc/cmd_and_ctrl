package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fear of Being Hunted — Enchantment Creature — Nightmare {1}{R}{R}, 4/2
// (Duskmourn: House of Horror, #4466):
//
//	"Haste
//	 This creature must be blocked if able."
//
// Gaea's Protector's block requirement (MustBeBlocked, CR 509.1c) on a
// hasty red attacker: the defending player must block it with at least
// one creature that can.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a23293cf-4352-4ce6-8a68-3b30f908e137",
		Name:            "Fear of Being Hunted",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static:          []game.StaticAbility{MustBeBlocked()},
	})
}
