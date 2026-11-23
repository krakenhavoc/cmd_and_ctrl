package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rhox Pikemaster — Creature — Rhino Soldier {2}{W}{W}, 3/3 (Magic 2010, #5585):
//
//	"First strike (This creature deals combat damage before creatures
//	 without first strike.)
//	 Other Soldier creatures you control have first strike."
//
// First strike rides PrintedKeywords; the grant is a lord's keyword half
// over the other Soldiers you control (TribalKeywordGrant).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8866d33e-8159-4e10-8b11-914c12a0918f",
		Name:            "Rhox Pikemaster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Soldier"}, Others: true, YoursOnly: true}, "first strike"),
		},
	})
}
