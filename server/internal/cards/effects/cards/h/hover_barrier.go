package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hover Barrier — Creature — Illusion Wall {2}{U}, 0/6 (Return to Ravnica, #5566):
//
//	"Defender, flying"
//
// The printed keywords ride PrintedKeywords; the engine enforces them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a09cce6f-c957-49d7-af99-a7b2bce786e8",
		Name:            "Hover Barrier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "flying"},
	})
}
