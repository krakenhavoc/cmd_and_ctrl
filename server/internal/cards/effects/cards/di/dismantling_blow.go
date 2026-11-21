package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:      "300cba5a-adbb-4852-be06-ac00d9d6fd37",
		Name:          "Dismantling Blow",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerRemove, Kicked: &game.Purpose{Draws: 2}},
		Targets:       effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{2}{U}")},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DestroyChosenPermanent(ctx.Game, ctx.Item)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.WasKicked() {
					return nil
				}
				return effects.DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
			}),
		),
	})
}
