package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deft Duelist — {W}{U} Creature — Human Rogue 2/1 (Shards of Alara, S82 OLD b515):
//
//	"First strike
//	 Shroud"
//
// Keywords only: each is read through the layered view.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2b6aaa8b-b90d-459f-95d4-9e00ee134f8e",
		Name:            "Deft Duelist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "shroud"},
	})
}
