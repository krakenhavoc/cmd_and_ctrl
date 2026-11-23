package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Innocuous Researcher — Creature — Centaur Detective {3}{G}, 3/4
// (Murders at Karlov Manor Commander):
//
//	"Parley — Whenever this creature attacks, each player reveals the top
//	 card of their library. For each nonland card revealed this way, you
//	 investigate. Then each player draws a card.
//	 At the beginning of your end step, you may untap all lands you
//	 control. If you do, you can't cast spells until your next turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef764497-800e-4cbd-805e-7635919a983c",
		Name:         "Innocuous Researcher",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks(innocuousResearcherLabel, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.Cmdr2S49Parley(g, ctx.Source(), innocuousResearcherLabel, func(nonland int) error {
					return effects.Investigate{N: nonland}.Apply(ctx)
				})
			}),
			effects.AtYourEndStep("Innocuous Researcher — you may untap all lands you control", func(g *game.Game, item *game.StackItem) error {
				return effects.MayChoice{
					Question: "Innocuous Researcher — untap all lands you control? (You can't cast spells until your next turn.)",
					Then:     innocuousResearcherAnswer,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

const innocuousResearcherLabel = "Innocuous Researcher — each player reveals the top card of their library; you investigate for each nonland card; then each player draws a card"

var innocuousResearcherAnswer = effects.MayYes("confirm/innocuous-researcher-untap", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if err := effects.UntapAllLandsControlledBy(ctx.Game, ctx.Controller(), ctx); err != nil {
		return err
	}
	return effects.RestrictCasting{
		Player:   ctx.Controller(),
		Rule:     game.CastBanRule{Kind: game.CastBanOutright},
		Label:    "Innocuous Researcher — you can't cast spells until your next turn",
		Duration: effects.DurationUntilYourNextTurn(ctx, ctx.Controller()),
	}.Apply(ctx)
})
