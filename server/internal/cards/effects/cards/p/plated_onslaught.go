package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plated Onslaught — Instant {3}{W}{W} (Phyrexia: All Will Be One):
//
//	"Affinity for artifacts
//	 Creatures you control get +2/+1 until end of turn."
//
// Affinity is the shared self cost modifier (CR 702.41a). The pump goes on
// the creatures you control as it resolves (CR 611.2c): one that arrives
// later is not boosted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1589d90d-411a-4e77-9771-51518b2b224b",
		Name:         "Plated Onslaught",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for artifacts", effects.Artifact()),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 2, Toughness: 1, Label: "Plated Onslaught — +2/+1"}.Apply(ctx)
		},
	})
}
