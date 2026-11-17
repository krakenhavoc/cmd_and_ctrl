package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Union of the Third Path — Instant for {2}{W} (The Brothers' War):
//
//	"Draw a card, then you gain life equal to the number of cards in your hand."
//
// The hand is counted after the draw has finished, so a replacement or a
// prompt the draw asks comes first (CR 608.2c).
func init() {
	p := game.Purpose{Draws: 1}
	Register(Spec{
		OracleID:     "4492c329-174a-44aa-be4d-9f018173b9a7",
		Name:         "Union of the Third Path",
		Purpose:      p,
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				pl := ctx.Game.PlayerByIDForEffect(ctx.Controller())
				if pl == nil {
					return nil
				}
				return GainLife{Player: ctx.Controller(), Amount: len(pl.Hand.Cards)}.Apply(ctx)
			}),
		),
	})
}
