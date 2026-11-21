package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "052e8e1c-d006-4c9f-90da-deafbbf7fed3",
		Name:         "Rankle's Prank",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOneOrMore(
			effects.ModeSteps("Each player discards two cards.", nil,
				effects.Instruction(func(ctx *effects.Context) error {
					ctx.Game.EachPlayerDiscardsForEffect(uuid.Nil, game.DiscardPrompt{
						Source: ctx.Source(), N: 2, Question: "Rankle's Prank — discard two cards",
					})
					return nil
				})),
			effects.ModeSteps("Each player loses 4 life.", nil,
				effects.Instruction(func(ctx *effects.Context) error {
					for _, id := range effects.TablePlayers(ctx) {
						if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), id, -4); err != nil {
							return err
						}
					}
					return nil
				})),
			effects.ModeSteps("Each player sacrifices two creatures of their choice.", nil,
				effects.Instruction(func(ctx *effects.Context) error {
					return effects.JmpB1EachPlayerSacrificesN(2, "Rankle's Prank — choose two creatures to sacrifice")(ctx.Game, ctx.Item)
				})),
		),
	})
}
