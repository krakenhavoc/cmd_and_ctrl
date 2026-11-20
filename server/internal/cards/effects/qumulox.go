package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Qumulox — Creature — Beast {6}{U}{U}, 5/4 (Fifth Dawn, #5604):
//
//	"Affinity for artifacts (This spell costs {1} less to cast for each
//	 artifact you control.)
//	 Flying"
//
// Broodstar's affinity for artifacts (the shared self cost modifier, CR
// 702.41a, generic mana only) on a flier.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cec11318-4578-4c66-a9be-1704df0c296c",
		Name:            "Qumulox",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for artifacts", Artifact()),
		},
	})
}
