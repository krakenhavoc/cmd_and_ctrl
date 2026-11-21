package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "1333d690-a6dc-4ef8-8c57-2ba16727c0a5",
		Name:            "Valkyrie Aerial Unit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for artifacts", effects.Artifact()),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Valkyrie Aerial Unit — surveil 2", effects.Do(effects.Surveil{N: 2})),
		},
	})
}
