package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Make It So — Sorcery {3}{U} (Star Trek, #4973):
//
//	"Choose one —
//	 • Untap target permanent you control. Target player draws two cards.
//	 • Create a 1/1 red Officer creature token and draw two cards."
//
// The first bullet has two target clauses, each judged on its own as the
// spell resolves (CR 608.2b): a permanent that has left is not untapped,
// and the player still draws. The second bullet is the token and then the
// draw, in printed order, the draw waiting for any prompt the token asks
// (CR 608.2c).
//
// No simplification.
func init() {
	draw := effects.Mode("Create a 1/1 red Officer creature token and draw two cards.")
	draw.Purpose = game.Purpose{Draws: 2, Tokens: 1, CreatureTokens: 1}
	effects.Register(effects.Spec{
		OracleID:     "a9368972-4697-4a10-97e3-324149a26fba",
		Name:         "Make It So",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Untap target permanent you control. Target player draws two cards.",
				effects.Clauses(effects.TargetPermanent("target permanent you control", effects.YouControl()),
					effects.TargetPlayer("target player"))),
			draw,
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.HasMode(0) {
					return nil
				}
				if t, ok := effects.ModeClauseTarget(ctx, 0, 0); ok && t.Kind == game.TargetCard {
					if err := (effects.UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.HasMode(0) {
					return nil
				}
				if t, ok := effects.ModeClauseTarget(ctx, 0, 1); ok && t.Kind == game.TargetPlayer {
					return effects.DrawCards{Player: t.ID, N: 2}.Apply(ctx)
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.HasMode(1) {
					return nil
				}
				return effects.CreateToken{Template: effects.TokenCard("1/1 red Officer"), N: 1}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.HasMode(1) {
					return nil
				}
				return effects.DrawCards{N: 2}.Apply(ctx)
			}),
		),
	})
}
