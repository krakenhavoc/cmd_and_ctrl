package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Daggerdrome Imp — Creature — Imp {1}{B}, 1/1 (Return to Ravnica, #5566):
//
//	"Flying / Lifelink (Damage dealt by this creature also causes you
//	 to gain that much life.)"
//
// The printed keywords ride PrintedKeywords; the engine enforces them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "894979ba-e4c7-4914-b0d3-133cbaed38c2",
		Name:            "Daggerdrome Imp",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
	})
}
