package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mournwillow — Creature — Plant Skeleton {1}{B}{G}, 3/2 (Eldritch Moon, S82
// OLD, #5541):
//
//	"Haste
//	 Delirium — When this creature enters, if there are four or more card
//	 types among cards in your graveyard, creatures with power 2 or less
//	 can't block this turn."
//
// An intervening "if" (CR 603.4): checked as the creature enters and again
// as the trigger resolves. Rumbling Ruin's can't-block set, here for every
// player's creatures (not only opponents'); power is read live, so a
// creature that arrives later this turn with power 2 or less can't block
// either (CR 611.2c, 509.1b).
//
// No simplification.
func init() {
	const name = "Mournwillow"
	effects.Register(effects.Spec{
		OracleID:        "742f6478-7c9e-43a4-81c1-09383fe65dad",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.YouHaveDeliriumWhen),
				name+" — Delirium: creatures with power 2 or less can't block this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if !effects.DrcDeliriumMet(g, ctx.Controller()) {
						return nil
					}
					return effects.CantBlockThisTurnMatching(false,
						name+" — creatures with power 2 or less can't block this turn",
						game.PermanentQuery{Types: []string{"creature"}}.WithPowerAtMost(2)).Apply(ctx)
				}),
		},
	})
}
