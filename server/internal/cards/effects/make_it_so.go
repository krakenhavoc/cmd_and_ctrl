package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	draw := Mode("Create a 1/1 red Officer creature token and draw two cards.")
	draw.Purpose = game.Purpose{Draws: 2, Tokens: 1, CreatureTokens: 1}
	Register(Spec{
		OracleID:     "a9368972-4697-4a10-97e3-324149a26fba",
		Name:         "Make It So",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Untap target permanent you control. Target player draws two cards.",
				Clauses(TargetPermanent("target permanent you control", YouControl()),
					TargetPlayer("target player"))),
			draw,
		),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				if !ctx.HasMode(0) {
					return nil
				}
				if t, ok := ModeClauseTarget(ctx, 0, 0); ok && t.Kind == game.TargetCard {
					if err := (UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
			Instruction(func(ctx *Context) error {
				if !ctx.HasMode(0) {
					return nil
				}
				if t, ok := ModeClauseTarget(ctx, 0, 1); ok && t.Kind == game.TargetPlayer {
					return DrawCards{Player: t.ID, N: 2}.Apply(ctx)
				}
				return nil
			}),
			Instruction(func(ctx *Context) error {
				if !ctx.HasMode(1) {
					return nil
				}
				return CreateToken{Template: TokenCard("1/1 red Officer"), N: 1}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				if !ctx.HasMode(1) {
					return nil
				}
				return DrawCards{N: 2}.Apply(ctx)
			}),
		),
	})
}
