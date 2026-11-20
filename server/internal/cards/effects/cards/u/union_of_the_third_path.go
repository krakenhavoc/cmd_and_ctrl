package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Union of the Third Path — Instant for {2}{W} (The Brothers' War):
//
//	"Draw a card, then you gain life equal to the number of cards in your hand."
//
// The hand is counted after the draw has finished, so a replacement or a
// prompt the draw asks comes first (CR 608.2c).
func init() {
	p := game.Purpose{Draws: 1}
	effects.Register(effects.Spec{
		OracleID:     "4492c329-174a-44aa-be4d-9f018173b9a7",
		Name:         "Union of the Third Path",
		Purpose:      p,
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				pl := ctx.Game.PlayerByIDForEffect(ctx.Controller())
				if pl == nil {
					return nil
				}
				return effects.GainLife{Player: ctx.Controller(), Amount: len(pl.Hand.Cards)}.Apply(ctx)
			}),
		),
	})
}
