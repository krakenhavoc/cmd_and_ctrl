package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gloomhunter — {2}{B} Creature — Bat 2/1 (Rise of the Eldrazi, S82 OLD #5581):
//
//	Flying

func init() {
	effects.Register(effects.Spec{
		OracleID:        "637b1a19-2925-422e-b867-f87537803071",
		Name:            "Gloomhunter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
