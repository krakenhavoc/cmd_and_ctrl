package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Voyage Home — Sorcery {5}{W}{U}:
//
//	"Affinity for artifacts (This spell costs {1} less to cast for
//	 each artifact you control.)
//	 You draw three cards and gain 3 life."
//
// #746: affinity through Spec.SelfCostModifiers (CR 702.41a).
func init() {
	Register(Spec{
		OracleID:     "5c0cbc44-6c31-44fe-a3da-a97a37a01726",
		Name:         "Voyage Home",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for artifacts", Artifact()),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (DrawCards{Player: ctx.Controller(), N: 3}).Apply(ctx); err != nil {
				return err
			}
			return GainLife{Player: ctx.Controller(), Amount: 3}.Apply(ctx)
		},
	})
}
