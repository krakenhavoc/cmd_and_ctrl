package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gaea's Blessing — Sorcery {1}{G} (Secret Lair Drop):
//
//	"Target player shuffles up to three target cards from their graveyard into
//	 their library.
//	 Draw a card.
//	 When this card is put into your graveyard from your library, shuffle your
//	 graveyard into your library."
//
// Krosan Reclamation's two clauses at three cards (a player, then
// FromTargetPlayersGraveyard cards tied to it) and the shared
// TargetPlayerShufflesTheirCards body, then the caster draws. The trigger is
// Narcomoeba's: it works from the graveyard the card has just landed in, whether
// it was milled, surveilled or put there by a search, and shuffles the whole
// graveyard (the card with it) into its owner's library, Finale of Revelation's
// shuffle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd5478df-3854-404b-8634-9debcef69b99",
		Name:         "Gaea's Blessing",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Answers: game.AnswerValue},
		Targets: effects.Clauses(
			effects.TargetPlayer("target player"),
			effects.FromTargetPlayersGraveyard(
				effects.TargetCardInGraveyard("up to three target cards from their graveyard").WithCount(0, 3), 0),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.TargetPlayerShufflesTheirCards(ctx, item.Targets, 0, 1); err != nil {
				return err
			}
			return effects.DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisIsPutIntoYourGraveyardFromYourLibrary(
				"Gaea's Blessing — shuffle your graveyard into your library",
				func(g *game.Game, item *game.StackItem) error {
					return effects.FinaleShuffleGraveyardIntoLibrary(effects.NewContext(g, item), item.Controller)
				}),
		},
	})
}
