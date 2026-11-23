package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blood Glutton — Creature — Vampire {4}{B}, 4/3 (Core Set 2021, #5511):
//
//	"Lifelink"
//
// The keyword rides PrintedKeywords; there is no other text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "065b8e5e-874e-4de3-9f12-41e4b1604561",
		Name:            "Blood Glutton",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
}
