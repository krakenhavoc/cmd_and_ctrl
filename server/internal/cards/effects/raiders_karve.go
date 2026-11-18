package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "1e9b3c9e-07ae-4de5-8607-93773825f655",
		Name:         "Raiders' Karve",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Raiders' Karve — look at the top card; if it's a land card, you may put it onto the battlefield tapped",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return PutFromLibraryOntoBattlefield{
						Player:   item.Controller,
						Cards:    g.LookAtTopOfLibraryForEffect(item.Controller, 1),
						Match:    Land(),
						Max:      1,
						Optional: true,
						Tapped:   true,
						Label:    "Raiders' Karve — you may put the land onto the battlefield tapped",
					}.Apply(ctx)
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    CrewCost(3),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Raiders' Karve"),
		}},
	})
}
