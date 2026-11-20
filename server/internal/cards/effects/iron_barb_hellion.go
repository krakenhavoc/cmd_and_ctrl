package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Iron-Barb Hellion — Creature — Hellion Beast {5}{R}, 5/4 (Fifth Dawn, #5604):
//
//	"Haste
//	 This creature can't block."
//
// Haste, and Fallen Askari's "can't block" restriction on itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6ec9f9a9-9e66-4d81-ae22-dc0de596acd5",
		Name:            "Iron-Barb Hellion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static:          []game.StaticAbility{RestrictSelf(game.CantBlock)},
	})
}
