package effects

import (
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
	Register(Spec{
		OracleID:     "ef764497-800e-4cbd-805e-7635919a983c",
		Name:         "Innocuous Researcher",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks(innocuousResearcherLabel, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return cmdr2S49Parley(g, ctx.Source(), innocuousResearcherLabel, func(nonland int) error {
					return Investigate{N: nonland}.Apply(ctx)
				})
			}),
			AtYourEndStep("Innocuous Researcher — you may untap all lands you control", func(g *game.Game, item *game.StackItem) error {
				return MayChoice{
					Question: "Innocuous Researcher — untap all lands you control? (You can't cast spells until your next turn.)",
					Then:     innocuousResearcherAnswer,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}

const innocuousResearcherLabel = "Innocuous Researcher — each player reveals the top card of their library; you investigate for each nonland card; then each player draws a card"

var innocuousResearcherAnswer = MayYes("confirm/innocuous-researcher-untap", func(ctx *Context, r game.ConfirmAnswer) error {
	if err := untapAllLandsControlledBy(ctx.Game, ctx.Controller(), ctx); err != nil {
		return err
	}
	return RestrictCasting{
		Player:   ctx.Controller(),
		Rule:     game.CastBanRule{Kind: game.CastBanOutright},
		Label:    "Innocuous Researcher — you can't cast spells until your next turn",
		Duration: DurationUntilYourNextTurn(ctx, ctx.Controller()),
	}.Apply(ctx)
})
