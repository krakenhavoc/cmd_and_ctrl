package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "ae92942b-919c-4ea9-b693-85fcef765d5a",
		Name:         "Fire",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny().WithCount(1, 2).Dividing(effects.Divide(2)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DealDividedDamage(ctx)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "ae92942b-919c-4ea9-b693-85fcef765d5a#1",
		Name:         "Ice",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target permanent"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.TapChosenPermanent(ctx.Game, ctx.Item)
			}),
			effects.DrawCards{N: 1},
		),
	})
}
