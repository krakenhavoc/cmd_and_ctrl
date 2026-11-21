package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Magnigoth Sentry — Creature — Treefolk {3}{G}, 4/4 (Foundations,
// #4465):
//
//	"Reach"
//
// Giant Spider's shape: reach rides PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b52824f9-7897-46b7-a4e5-a0261861cb5b",
		Name:            "Magnigoth Sentry",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
