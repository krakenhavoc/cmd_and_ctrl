package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thoughtcast — Sorcery {4}{U}:
//
//	"Affinity for artifacts (This spell costs {1} less to cast for
//	 each artifact you control.)
//	 Draw two cards."
//
// #746: affinity through Spec.SelfCostModifiers (CR 702.41a).
func init() {
	Register(Spec{
		OracleID:     "cce9bbff-82dc-4b2f-addd-d6715588de20",
		Name:         "Thoughtcast",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for artifacts", Artifact()),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
