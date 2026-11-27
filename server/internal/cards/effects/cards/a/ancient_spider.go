package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ancient Spider — Creature — Spider {2}{G}{W}, 2/5 (Planeshift, S82 OLD b643, #5615):
//
//	"Reach
//	 First strike"
//
// Both words are printed engine keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "28537019-3061-4efc-8e87-9bb805bf520f",
		Name:            "Ancient Spider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "first strike"},
	})
}
