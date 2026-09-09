package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blasphemous Act — Sorcery {8}{R}:
//
//	"This spell costs {1} less to cast for each creature on the
//	battlefield. Destroy all creatures."
//
// Sandbox simplification: THE COST REDUCTION IS NOT IMPLEMENTED —
// this casts at its printed {8}{R}. Cost modification is S28
// territory (Spec has no hook for it), and faking it by lowering the
// printed cost would be wrong in the other direction. The board
// wipe itself is complete and correct.
//
// Iterates a snapshot of CreatureIDs rather than the live
// battlefield: DestroyTarget moves cards out from under the range as
// it goes, and a dies-trigger resolving mid-loop could add one.
func init() {
	Register(Spec{
		OracleID: "7a2484a9-04fd-41a0-8224-610c1c07ed10",
		Name:     "Blasphemous Act",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, id := range ctx.CreatureIDs() {
				if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
