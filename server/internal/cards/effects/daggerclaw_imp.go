package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Daggerclaw Imp — Creature — Imp {2}{B}, 3/1 (Ravnica: Clue Edition):
//
//	"Flying
//	 This creature can't block."
//
// Flying is the engine keyword; "can't block" is Craven Giant's static
// restriction on itself (CR 509.1b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "689c5a47-24ec-4cbf-bc93-ebb492732cad",
		Name:            "Daggerclaw Imp",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{RestrictSelf(game.CantBlock)},
	})
}
