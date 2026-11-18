package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// White Sun's Zenith — Instant {X}{W}{W}{W}:
//
//	"Create X 2/2 white Cat creature tokens. Shuffle White Sun's Zenith
//	 into its owner's library."
//
// Blue Sun's Zenith's tuck-and-shuffle (the resolving card moves itself, CR
// 608.2n) after X Cat tokens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ea4a9299-c2b3-4092-89dd-28142b203d60",
		Name:         "White Sun's Zenith",
		Completeness: CompletenessFull,
		XMatters:     true,
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return CreateToken{Controller: ctx.Controller(), Template: TokenCard("2/2 white Cat"), N: ctx.X()}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				return ctx.Game.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{ctx.Source()}, game.TuckOptions{}, game.KeyedTail{Ref: thenShuffle, Player: ctx.Item.Owner})
			}),
		),
	})
}
