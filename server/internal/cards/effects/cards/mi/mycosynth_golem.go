package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mycosynth Golem — Artifact Creature — Golem {11}, 4/5:
//
//	"Affinity for artifacts (This spell costs {1} less to cast for
//	 each artifact you control.)
//	 Artifact creature spells you cast have affinity for artifacts."
//
// #746: both slots — the ADR's own example of a card that needs both.
// The grant is a battlefield cost modifier, one instance per Golem
// (CR 702.41b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ebcd864a-b7dd-4330-89e1-80576a9437ec",
		Name:         "Mycosynth Golem",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for artifacts", effects.Artifact()),
		},
		CostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.PermanentsYouControl(effects.Artifact()),
				"Artifact creature spells you cast have affinity for artifacts.",
				effects.YourSpell(), effects.ArtifactCreatureSpell()),
		},
	})
}
