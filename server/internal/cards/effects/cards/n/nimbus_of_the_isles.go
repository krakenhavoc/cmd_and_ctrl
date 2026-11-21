package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nimbus of the Isles — Creature — Elemental {4}{U}, 3/3 (Magic 2015,
// #5556):
//
//	"Flying"
//
// Armored Pegasus's printed flying, carried as data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3d529824-8e82-4c44-8e28-ed3419740c11",
		Name:            "Nimbus of the Isles",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
