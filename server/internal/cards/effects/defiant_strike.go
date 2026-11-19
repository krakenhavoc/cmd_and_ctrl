package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Defiant Strike — Instant {W} (Foundations Jumpstart, #5929):
//
//	"Target creature gets +1/+0 until end of turn.
//	 Draw a card."
//
// The pump and then the draw; a target that is gone as the spell resolves
// fizzles the whole spell, draw included (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f8c2d9c5-6fe6-4698-9972-83cdd625d94b",
		Name:         "Defiant Strike",
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 0)),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if err := (BoostUntilEOT{Target: item.Targets[0].ID, Power: 1, Label: "Defiant Strike — +1/+0"}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
