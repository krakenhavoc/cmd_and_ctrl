package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rankle's Prank — Sorcery {2}{B}{B} (Wilds of Eldraine, #4471):
//
//	"Choose one or more —
//	 • Each player discards two cards.
//	 • Each player loses 4 life.
//	 • Each player sacrifices two creatures of their choice."
//
// A modal spell whose chosen bullets resolve in printed order, each
// waiting for the prompts the one before asked (ModeSteps, CR 608.2c).
// The discard is a real choice over each hand in APNAP order (Garruk's
// fan-out, no payout clause); the life loss is every living player,
// the caster included; the sacrifice is Barter in Blood's one choice of
// two creatures per player, taken together (CR 701.21a), a player with
// fewer sacrificing what they have.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "052e8e1c-d006-4c9f-90da-deafbbf7fed3",
		Name:         "Rankle's Prank",
		Completeness: CompletenessFull,
		Modes: ChooseOneOrMore(
			ModeSteps("Each player discards two cards.", nil,
				Instruction(func(ctx *Context) error {
					ctx.Game.EachPlayerDiscardsForEffect(uuid.Nil, game.DiscardPrompt{
						Source: ctx.Source(), N: 2, Question: "Rankle's Prank — discard two cards",
					})
					return nil
				})),
			ModeSteps("Each player loses 4 life.", nil,
				Instruction(func(ctx *Context) error {
					for _, id := range tablePlayers(ctx) {
						if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), id, -4); err != nil {
							return err
						}
					}
					return nil
				})),
			ModeSteps("Each player sacrifices two creatures of their choice.", nil,
				Instruction(func(ctx *Context) error {
					return jmpB1EachPlayerSacrificesN(2, "Rankle's Prank — choose two creatures to sacrifice")(ctx.Game, ctx.Item)
				})),
		),
	})
}
