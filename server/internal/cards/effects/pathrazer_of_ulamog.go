package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pathrazer of Ulamog — Creature — Eldrazi {11}, 9/9:
//
//	"Annihilator 3
//	 This creature can't be blocked except by three or more creatures."
//
// Annihilator 3 is the canonical keyword token, and the engine makes
// its attack trigger (ADR 0113 §2). The blocking rule is the minimum
// count rule Rampaging Ceratops uses: one or two blockers are refused
// at declaration.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4bf95747-4572-49b8-b892-87fe7f910252",
		Name:            "Pathrazer of Ulamog",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"annihilator 3"},
		BlockRules:      []game.BlockRule{MinBlockers(OnSelf(), 3)},
	})
}
