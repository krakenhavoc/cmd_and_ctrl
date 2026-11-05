package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Utrom Monitor — Artifact Creature — Utrom Scientist {4}{U}, 3/3 (TMNT
// Commander, #5011):
//
//	"Affinity for artifacts
//	 Flying"
//
// Affinity is the card's own cost modifier (CR 702.41a): {1} less for
// each artifact you control, off the generic part only, its mana value
// unchanged. Flying rides PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "976e1fe6-ac32-4ea9-9828-acd3c4791e01",
		Name:            "Utrom Monitor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for artifacts", Artifact()),
		},
	})
}
