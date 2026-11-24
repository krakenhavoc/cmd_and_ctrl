package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ethersworn Sphinx — Artifact Creature — Sphinx {7}{W}{U}, 4/4 (Modern
// Horizons 2, #5501):
//
//	"Affinity for artifacts
//	 Flying
//	 Cascade"
//
// Affinity is the self cost reduction of CR 702.41a over the artifacts its
// caster controls; cascade is the shared keyword trigger (CR 702.85a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "71366469-04ea-4526-bf24-7abc02fdc5ab",
		Name:            "Ethersworn Sphinx",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for artifacts", effects.Artifact()),
		},
		Triggered: []game.TriggeredAbility{effects.Cascade()},
	})
}
