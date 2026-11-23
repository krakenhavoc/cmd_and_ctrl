package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trained Caracal — Creature — Cat {W}, 1/1 (Return to Ravnica, #5566):
//
//	"Lifelink (Damage dealt by this creature also causes you to gain
//	 that much life.)"
//
// The printed keywords ride PrintedKeywords; the engine enforces them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a939f01b-afce-4d92-a897-1a6dc3d788c6",
		Name:            "Trained Caracal",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
}
