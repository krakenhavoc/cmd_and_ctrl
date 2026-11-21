package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oxidda Golem — Artifact Creature — Golem {6} (Darksteel, #5605):
//
//	"Affinity for Mountains (This spell costs {1} less to cast for each
//	 Mountain you control.)
//	 Haste"
//
// Spire Golem's shape: affinity is the shared self cost modifier counting
// the Mountains its caster controls (CR 702.41a), and haste is a printed
// keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9c51d843-d155-44ed-ab1f-dfa2ff848f53",
		Name:            "Oxidda Golem",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for Mountains", effects.And(effects.Land(), effects.Subtype("Mountain"))),
		},
	})
}
