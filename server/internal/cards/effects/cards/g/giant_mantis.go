package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Giant Mantis — Creature — Insect {3}{G}, 2/4 (Battle for Zendikar, #5547):
//
//	"Reach"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "481ede6a-3ac3-45bc-85aa-311b3dbefc40",
		Name:            "Giant Mantis",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
