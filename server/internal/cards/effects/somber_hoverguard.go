package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Somber Hoverguard — Creature — Drone {5}{U}, 3/2 (Mirrodin, #5606):
//
//	"Affinity for artifacts (This spell costs {1} less to cast for each
//	 artifact you control.)
//	 Flying"
//
// Frogmite's affinity, the shared self cost modifier (CR 702.41a),
// generic mana only, so the {U} is always paid; and printed flying.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3c88108e-d983-427a-9b85-b9cdd2d5083b",
		Name:            "Somber Hoverguard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for artifacts", Artifact()),
		},
	})
}
