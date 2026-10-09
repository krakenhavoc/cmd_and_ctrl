package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Terminal Criticism — Instant {1}{B} (Reality Fracture, tracker #2795):
//
//	"Destroy target creature or planeswalker that's blue or red. You gain
//	 1 life."
//
// The colour clause is a target predicate over the permanent's effective
// colours, so a creature that stopped being blue in response is no longer
// a legal target and the spell does nothing (CR 608.2b). A multicoloured
// permanent that is blue or red passes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "213b0814-f768-4408-9cb7-6f5960bcbb7a",
		Name:         "Terminal Criticism",
		Completeness: CompletenessFull,
		Targets: TargetPermanent("target creature or planeswalker that's blue or red",
			And(Or(Creature(), Planeswalker()), Or(OfColor("U"), OfColor("R")))),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if err := (DestroyTarget{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
			return GainLife{Player: ctx.Controller(), Amount: 1}.Apply(ctx)
		},
	})
}
