package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tower Gargoyle — Artifact Creature — Gargoyle {1}{W}{U}{B}, 4/4 (Alara
// Reborn):
//
//	"Flying"
//
// Flying is a printed keyword; there is no other text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "681ca1eb-2ef5-4d60-b834-58bd9b5a47b8",
		Name:            "Tower Gargoyle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
