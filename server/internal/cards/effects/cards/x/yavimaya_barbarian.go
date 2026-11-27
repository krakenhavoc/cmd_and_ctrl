package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yavimaya Barbarian — Creature — Elf Barbarian {R}{G}, 2/2 (Invasion, S82 #5616):
//
//	"Protection from blue"
//
// Protection from blue is a printed keyword: it can't be blocked, targeted, dealt damage or enchanted by anything blue (CR 702.16).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "04576b4d-2977-4cdf-b4a5-b0b47995e962",
		Name:            "Yavimaya Barbarian",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from blue"},
	})
}
