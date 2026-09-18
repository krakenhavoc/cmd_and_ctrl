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
// SANDBOX GAP, weaker than printed: the "choose both" rider is not
// offered. Each mode carries its own target, and a modal spec whose
// maximum is above one may carry a target on at most one option
// (per-mode target slots are the open multi-target work), so the
// card is "choose one" whether or not a commander is on the board.
// Never stronger: the wider choice is simply absent.
func init() {
	Register(Spec{
		OracleID:     "d3cec4b5-bc93-44a2-a29d-3478f0a5dac6",
		Name:         "Drown in Dreams",
		XMatters:     true,
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Choosing both modes when you control a commander isn't implemented — you always choose one."},
		Modes: ChooseOne(
			Mode("Target player draws X cards.", TargetPlayer("target player")),
			Mode("Target player mills twice X cards.", TargetPlayer("target player")),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			player := item.Targets[0].ID
			if ctx.HasMode(0) {
				return DrawCards{Player: player, N: ctx.X()}.Apply(ctx)
			}
			if ctx.HasMode(1) {
				return MillCards{Player: player, N: 2 * ctx.X()}.Apply(ctx)
			}
			return nil
		},
	})
}
