package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Branchblight Stalker — Creature — Phyrexian Elf Scout {1}{G}, 3/1 (Phyrexia: All Will Be One, #5479):
//
//	"Toxic 2"
//
// Every printed keyword is read through the layered view. Toxic is the
// engine's damage tail (infect_wither_toxic.go, CR 702.164).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ae685a9b-4cb8-4a8a-935d-378102431c51",
		Name:            "Branchblight Stalker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"toxic 2"},
	})
}
