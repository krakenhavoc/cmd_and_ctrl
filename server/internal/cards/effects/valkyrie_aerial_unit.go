package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Valkyrie Aerial Unit — Artifact Creature — Construct {5}{U}{U}, 5/4
// (Final Fantasy, #4462):
//
//	"Affinity for artifacts (This spell costs {1} less to cast for each
//	 artifact you control.)
//	 Flying
//	 When this creature enters, surveil 2."
//
// Thought Monitor's affinity (CR 702.41a) and A.I.M. Synthoids' enters
// trigger: the surveil goes on the stack, so it can be answered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1333d690-a6dc-4ef8-8c57-2ba16727c0a5",
		Name:            "Valkyrie Aerial Unit",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for artifacts", Artifact()),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Valkyrie Aerial Unit — surveil 2", Do(Surveil{N: 2})),
		},
	})
}
