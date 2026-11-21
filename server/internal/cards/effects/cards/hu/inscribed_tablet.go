package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inscribed Tablet — Artifact {1} (Dominaria United, S82 DMU batch 4):
//
//	"{1}, {T}, Sacrifice this artifact: Reveal the top five cards of your
//	 library. Put a land card from among them into your hand and the rest
//	 on the bottom of your library in a random order. If you didn't put a
//	 card into your hand this way, draw a card."
//
// Grisly Salvage's reveal of the top five (every player sees them), then a
// mandatory pick of a land (no "may"). The rest go to the bottom in a random
// order, and with no land among the five (nothing went to the hand) the
// controller draws.
//
// No simplification.
func init() {
	const label = "Inscribed Tablet — reveal the top five cards, put a land card from among them into your hand and the rest on the bottom in a random order; if you didn't, draw a card"
	effects.Register(effects.Spec{
		OracleID:     "399d19cd-e6fa-4959-8d2a-9f1cfd52dce7",
		Name:         "Inscribed Tablet",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice this artifact: " + label,
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				player := ctx.Controller()
				return effects.TakeFromLibraryToHand{
					Player: player,
					Cards:  g.RevealTopOfLibraryForEffect(player, ctx.Source(), 5, label),
					Match:  effects.Land(),
					Max:    1,
					Label:  label,
					Key:    effects.DmuB04TabletTake,
				}.Apply(ctx)
			},
		}},
	})
}
