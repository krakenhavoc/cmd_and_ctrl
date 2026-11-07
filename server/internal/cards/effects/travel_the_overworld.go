package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Travel the Overworld — Sorcery {5}{U}{U} (Final Fantasy, #4462):
//
//	"Affinity for Towns (This spell costs {1} less to cast for each Town
//	 you control.)
//	 Draw four cards."
//
// Affinity (CR 702.41a) is Thought Monitor's self cost modifier over
// Towns: it reduces only the generic part, so the spell never costs
// less than {U}{U}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "61e3709e-7ecd-4a59-bca1-019d16795740",
		Name:         "Travel the Overworld",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for Towns", HasSubtype("Town")),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DrawCards{N: 4}.Apply(ctx)
		},
	})
}
