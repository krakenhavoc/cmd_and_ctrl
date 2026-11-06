package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Memory Guardian — Artifact Creature — Robot Artificer {4}{U}, 3/4
// (Aetherdrift, #4464):
//
//	"Affinity for artifacts (This spell costs {1} less to cast for each
//	 artifact you control.)
//	 Flying"
//
// Myr Enforcer's affinity (CR 702.41a), a cost modifier on the card
// itself that counts the artifacts its caster controls, with flying as
// the printed keyword. The reduction is generic only, so {U} is always
// paid.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "dac33359-0317-4b11-9078-fccfd2321e53",
		Name:            "Memory Guardian",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for artifacts", Artifact()),
		},
	})
}
