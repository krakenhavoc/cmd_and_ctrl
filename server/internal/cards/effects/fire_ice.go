package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fire // Ice — split card (Apocalypse, #5614):
//
//	Fire — Instant {1}{R}: "Fire deals 2 damage divided as you choose
//	 among one or two targets."
//	Ice — Instant {1}{U}: "Tap target permanent.
//	 Draw a card."
//
// Each half is an ADR 0034 face (ADR 0103): Fire under the bare oracle
// ID, Ice under "#1". Only one half can be cast (CR 709.3; the card has
// no fuse). Fire is Twin Bolt's division (CR 601.2d: each of the one or
// two targets gets at least 1, the shares summing to 2). Ice is
// tapChosenPermanent and then a draw, two steps (ResolveSteps, CR
// 608.2c); an Ice whose only target is gone does not resolve at all
// (CR 608.2b), so it draws nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ae92942b-919c-4ea9-b693-85fcef765d5a",
		Name:         "Fire",
		Completeness: CompletenessFull,
		Targets:      TargetAny().WithCount(1, 2).Dividing(Divide(2)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DealDividedDamage(ctx)
		},
	})
	Register(Spec{
		OracleID:     "ae92942b-919c-4ea9-b693-85fcef765d5a#1",
		Name:         "Ice",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target permanent"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return tapChosenPermanent(ctx.Game, ctx.Item)
			}),
			DrawCards{N: 1},
		),
	})
}
