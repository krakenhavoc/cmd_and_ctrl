package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Underworld Cookbook — Artifact — Book {1} (Modern Horizons 2, #5501):
//
//	"{T}, Discard a card: Create a Food token.
//	 {4}, {T}, Sacrifice this artifact: Return target creature card from your
//	 graveyard to your hand."
//
// The discard is a cost, so it is paid before the ability goes on the stack
// and a card with no hand to pay it cannot activate. The second ability's
// target is a creature card in its controller's graveyard; if it left in
// response the ability does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "37d5c720-d090-44bc-91df-0a7bab79a03e",
		Name:         "The Underworld Cookbook",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}, Discard a card: Create a Food token.",
				Purpose: game.Purpose{Tokens: 1, Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.TapCost(), effects.DiscardACard()),
				Effect:  effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1}),
			},
			{
				Label:   "{4}, {T}, Sacrifice this artifact: Return target creature card from your graveyard to your hand.",
				Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost(), effects.SacrificeThis()),
				Targets: effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
				Effect:  effects.ReturnFirstLegalGraveyardTargetToHand,
			},
		},
	})
}
