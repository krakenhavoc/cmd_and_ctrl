package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "ea4a9299-c2b3-4092-89dd-28142b203d60",
		Name:         "White Sun's Zenith",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("2/2 white Cat"), N: ctx.X()}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return ctx.Game.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{ctx.Source()}, game.TuckOptions{}, game.KeyedTail{Ref: effects.ThenShuffle, Player: ctx.Item.Owner})
			}),
		),
	})
}
