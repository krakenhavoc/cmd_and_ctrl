package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kruphix's Insight — Sorcery {2}{G}:
//
//	"Reveal the top six cards of your library. Put up to three enchantment
//	 cards from among them into your hand and the rest of the revealed cards
//	 into your graveyard."
//
// The six cards are revealed to every player (CR 701.20a), then the
// controller picks up to three enchantment cards (the prompt's floor is
// zero); the rest are put, not milled, into the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fadfbceb-4392-4e7e-be29-cbf372135d16",
		Name:         "Kruphix's Insight",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 3},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			label := "Kruphix's Insight — put up to three enchantment cards into your hand and the rest into your graveyard"
			player := ctx.Controller()
			revealed := ctx.Game.RevealTopOfLibraryForEffect(player, ctx.Source(), 6, label)
			return effects.TakeFromLibraryToHand{
				Player:   player,
				Cards:    revealed,
				Match:    effects.Enchantment(),
				Max:      3,
				Optional: true,
				Label:    label,
				Key:      effects.TakeRestGraveyard,
			}.Apply(ctx)
		},
	})
}
