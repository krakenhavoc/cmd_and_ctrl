package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raiders' Karve — Artifact — Vehicle {3}, 4/4 (Reprint, #5908):
//
//	"Whenever this Vehicle attacks, look at the top card of your library.
//	 If it's a land card, you may put it onto the battlefield tapped.
//	 Crew 3"
//
// Risen Reef's look: the top card is looked at and, if it is a land, the
// controller may put it onto the battlefield tapped. A card that stays is
// left on top of the library, as printed. Crew 3 is the ordinary CR 702.122
// crew ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1e9b3c9e-07ae-4de5-8607-93773825f655",
		Name:         "Raiders' Karve",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Raiders' Karve — look at the top card; if it's a land card, you may put it onto the battlefield tapped",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.PutFromLibraryOntoBattlefield{
						Player:   item.Controller,
						Cards:    g.LookAtTopOfLibraryForEffect(item.Controller, 1),
						Match:    effects.Land(),
						Max:      1,
						Optional: true,
						Tapped:   true,
						Label:    "Raiders' Karve — you may put the land onto the battlefield tapped",
					}.Apply(ctx)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    effects.CrewCost(3),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Raiders' Karve"),
		}},
	})
}
