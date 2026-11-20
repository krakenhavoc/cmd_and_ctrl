package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reckless Brute — Creature — Ogre Warrior {2}{R}, 3/1 (Magic 2013, #5568):
//
//	"Haste (This creature can attack and {T} as soon as it comes under your
//	 control.)
//	 This creature attacks each combat if able."
//
// Bloodrock Cyclops's attack requirement (CR 508.1d) on a body with
// haste, so it must attack the turn it arrives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5ba6a472-b2f3-4d8e-838c-269aaa5eb2e3",
		Name:            "Reckless Brute",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static:          []game.StaticAbility{AttacksEachCombat()},
	})
}
