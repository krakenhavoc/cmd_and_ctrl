package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Triumphant Reckoning — Sorcery {6}{W}{W}{W} (Secret Lair Drop):
//
//	"Return all artifact, enchantment, and planeswalker cards from your
//	 graveyard to the battlefield."
//
// Replenish's shape (one entry, so every enter trigger sees the whole batch,
// CR 603.6a) over a graveyard filter; the cards are the caster's own, so they
// come back under their control. A card that is none of the three stays.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0d3e0b38-a9ff-489b-8c00-4ca89deb4869",
		Name:         "Triumphant Reckoning",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerGraveyard},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ids := effects.GraveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool {
				return !c.IsToken() && (c.IsArtifact() || c.HasCardType("enchantment") || c.HasCardType("planeswalker"))
			})
			if len(ids) == 0 {
				return nil
			}
			return effects.ReturnFromGraveyardTogether{Targets: ids, Controller: ctx.Controller()}.Apply(ctx)
		},
	})
}
