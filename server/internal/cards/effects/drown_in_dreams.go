package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drown in Dreams — Instant {X}{2}{U} (EDHREC rank 2089):
//
//	"Choose one. If you control a commander as you cast this spell,
//	 you may choose both instead.
//	 • Target player draws X cards.
//	 • Target player mills twice X cards."
//
// Blue Sun's Zenith or a mill finisher, and with a commander out,
// both. Each mode is an ordinary X-scaled primitive on a player
// target; X rides the cast (ctx.X()).
//
// The commander rider is the conditional mode count #1590 built:
// `OrUpToIf(2, YouControlACommander)`, read at announce (CR 601.2b)
// and fixed from then on. With both chosen, each bullet has its own
// "target player" (#764's per-mode target groups), so the draw and
// the mill may point at different players; OptionTargets reads each
// bullet's own group. The body runs the bullets in PRINTED order (CR
// 608.2c) whatever order they were clicked in, which is observable
// here: drawing X and then milling 2X off the same player is not the
// same as the other way round.
func init() {
	Register(Spec{
		OracleID:     "d3cec4b5-bc93-44a2-a29d-3478f0a5dac6",
		Name:         "Drown in Dreams",
		XMatters:     true,
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Target player draws X cards.", TargetPlayer("target player")),
			Mode("Target player mills twice X cards.", TargetPlayer("target player")),
		).OrUpToIf(2, YouControlACommander),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if ctx.HasMode(0) {
				for _, t := range OptionTargets(ctx, 0) {
					if err := (DrawCards{Player: t.ID, N: ctx.X()}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			if ctx.HasMode(1) {
				for _, t := range OptionTargets(ctx, 1) {
					if err := (MillCards{Player: t.ID, N: 2 * ctx.X()}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			return nil
		},
	})
}
