package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Panther Robot — Artifact Creature — Cat Robot {10}, 8/8 (Marvel Super
// Heroes Commander, #4455):
//
//	"Affinity for artifacts (This spell costs {1} less to cast for each
//	 artifact you control.)
//	 Reach, trample"
//
// Myr Enforcer's cost modifier (AffinityFor, CR 702.41a) and printed
// keywords the importer carries.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "89115ed5-95bf-4b81-b352-5481c3addbed",
		Name:            "Panther Robot",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for artifacts", effects.Artifact()),
		},
	})
}
