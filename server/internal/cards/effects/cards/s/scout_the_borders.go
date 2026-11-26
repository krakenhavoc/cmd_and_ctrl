package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scout the Borders — Sorcery {2}{G} (Khans of Tarkir):
//
//	"Reveal the top five cards of your library. You may put a creature or
//	 land card from among them into your hand. Put the rest into your
//	 graveyard."
//
// Wakanda Forever!'s take-one-and-the-rest-to-the-graveyard over a public
// reveal (everyone sees all five). Declining puts all five into the
// graveyard; the rest go there as a zone move, not a mill (CR 701.17a), so a
// graveyard replacement sees it. A library of fewer than five reveals what
// is there.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "08781f2a-0d43-495c-aac8-7ff7a2e5752f",
		Name:         "Scout the Borders",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			const label = "Scout the Borders — you may put a creature or land card from among them into your hand; the rest go into your graveyard"
			return effects.TakeFromLibraryToHand{
				Player:   item.Controller,
				Cards:    ctx.Game.RevealTopOfLibraryForEffect(item.Controller, ctx.Source(), 5, label),
				Match:    effects.Or(effects.Creature(), effects.Land()),
				Max:      1,
				Optional: true,
				Label:    label,
				Key:      effects.TakeRestGraveyard,
			}.Apply(ctx)
		},
	})
}
