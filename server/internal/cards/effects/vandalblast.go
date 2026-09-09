package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vandalblast — Sorcery {R}:
//
//	"Destroy target artifact you don't control.
//	 Overload {4}{R} (You may cast this spell for its overload cost.
//	 If you do, change 'target' to 'each'.)"
//
// Sandbox simplification: OVERLOAD IS NOT IMPLEMENTED. Only the
// one-mana single-target mode exists. Overload is an alternative cost
// that also rewrites the targeting clause, which needs both the S28
// alternative-cost work and a way to swap a TargetSpec at cast time —
// neither exists, and faking it as a second mode would let a player
// wipe the board for {R}.
//
// The "you don't control" restriction is real and enforced:
// OpponentControls means your own Sol Ring is never a legal target.
func init() {
	Register(Spec{
		OracleID: "3567c3c8-b3c7-45b7-935b-b1fdbc973720",
		Name:     "Vandalblast",
		Targets:  TargetPermanent("target artifact you don't control", And(Artifact(), OpponentControls())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
