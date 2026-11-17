package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dismantling Blow — Instant {2}{W} (Invasion, #5616):
//
//	"Kicker {2}{U} (You may pay an additional {2}{U} as you cast this
//	 spell.)
//	 Destroy target artifact or enchantment. If this spell was kicked,
//	 draw two cards."
//
// CR 702.33's Kicker, read back at resolution with ctx.WasKicked. The
// two sentences are two steps (ResolveSteps), so the draw waits for
// anything the destruction asks first (CR 608.2c). The draw is not
// gated on the destruction: an indestructible target survives and a
// kicked Blow still draws. A Blow whose only target is gone does not
// resolve at all (CR 608.2b), so it draws nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "300cba5a-adbb-4852-be06-ac00d9d6fd37",
		Name:          "Dismantling Blow",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerRemove, Kicked: &game.Purpose{Draws: 2}},
		Targets:       TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OptionalCosts: []game.AdditionalCost{Kicker("{2}{U}")},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return destroyChosenPermanent(ctx.Game, ctx.Item)
			}),
			Instruction(func(ctx *Context) error {
				if !ctx.WasKicked() {
					return nil
				}
				return DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
			}),
		),
	})
}
