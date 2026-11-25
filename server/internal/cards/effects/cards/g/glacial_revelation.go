package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glacial Revelation — Sorcery {2}{G} (Modern Horizons):
//
//	"Reveal the top six cards of your library. You may put any number of snow
//	 permanent cards from among them into your hand. Put the rest into your
//	 graveyard."
//
// The six cards are revealed, then the caster picks any number (including
// none) of the snow permanent cards among them, which go to hand; the rest
// are put into the graveyard, not milled (Wakanda Forever!'s take with the
// rest going to the graveyard). A snow instant or sorcery is not a permanent
// card and can only go to the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "92359752-74aa-499e-b2db-0342d7596691",
		Name:         "Glacial Revelation",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			revealed := ctx.Game.RevealTopOfLibraryForEffect(item.Controller, ctx.Source(), 6,
				"Glacial Revelation — revealed from the top of the library")
			return effects.TakeFromLibraryToHand{
				Player:   item.Controller,
				Cards:    revealed,
				Match:    effects.And(effects.Permanent(), effects.Snow()),
				Max:      0,
				Optional: true,
				Label:    "Glacial Revelation — put any number of snow permanent cards from among them into your hand",
				Key:      effects.TakeRestGraveyard,
			}.Apply(ctx)
		},
	})
}
