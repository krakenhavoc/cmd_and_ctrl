package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Redress Fate — Sorcery {6}{W}{W} (Duskmourn Commander):
//
//	"Return all artifact and enchantment cards from your graveyard to the
//	 battlefield.
//	 Miracle {3}{W} (...)"
//
// The set is read once from the graveyard and the cards enter together
// (ReturnFromGraveyardTogether), so an Aura that returns picks its host as
// it enters (CR 303.4f) and a card that cannot enter stays put. Nothing is
// targeted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "e8ff6780-5f22-42fa-a7c1-7406e8203063",
		Name:             "Redress Fate",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Miracle("{3}{W}")},
		Purpose:          game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			ids := effects.GraveyardCardIDs(ctx, item.Controller, func(c game.Card) bool {
				return c.IsArtifact() || c.IsEnchantment()
			})
			if len(ids) == 0 {
				return nil
			}
			return effects.ReturnFromGraveyardTogether{Targets: ids, Controller: item.Controller}.Apply(ctx)
		},
	})
}
