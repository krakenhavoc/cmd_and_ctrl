package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orazca Puzzle-Door — Artifact {U} (The Lost Caverns of Ixalan, #4470):
//
//	"{1}, {T}, Sacrifice this artifact: Look at the top two cards of your
//	 library. Put one of those cards into your hand and the other into
//	 your graveyard."
//
// Sibsig Appraiser's look-and-take on an activated ability: a private
// look at the top two, a mandatory pick of one for the hand, and the
// other put into the graveyard (takeRestGraveyard), not discarded or
// milled. A library of one card puts that card into the hand. The
// sacrifice is part of the cost (CR 602.2b, CR 701.21a), so the door is
// gone before anyone can respond.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6bcf21ef-6641-4752-a0da-937a9277c812",
		Name:         "Orazca Puzzle-Door",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice this artifact: Look at the top two cards of your library. Put one of those cards into your hand and the other into your graveyard.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				player := ctx.Controller()
				return effects.TakeFromLibraryToHand{
					Player: player,
					Cards:  g.LookAtTopOfLibraryForEffect(player, 2),
					Max:    1,
					Label:  "Orazca Puzzle-Door — put one of them into your hand and the other into your graveyard",
					Key:    effects.TakeRestGraveyard,
				}.Apply(ctx)
			},
		}},
	})
}
