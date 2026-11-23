package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thoughtcast — Sorcery {4}{U}:
//
//	"Affinity for artifacts (This spell costs {1} less to cast for
//	 each artifact you control.)
//	 Draw two cards."
//
// #746: affinity through Spec.SelfCostModifiers (CR 702.41a).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cce9bbff-82dc-4b2f-addd-d6715588de20",
		Name:         "Thoughtcast",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for artifacts", effects.Artifact()),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
