package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Regalia — Legendary Artifact — Vehicle {4}, 4/4:
//
//	"Haste
//	 Whenever The Regalia attacks, reveal cards from the top of your
//	 library until you reveal a land card. Put that card onto the
//	 battlefield tapped and the rest on the bottom of your library in a
//	 random order.
//	 Crew 1"
//
// A Vehicle that ramps every time it swings. The attack trigger is the
// shared reveal-until sentence (RevealUntilThenPutOntoBattlefield,
// #745): the run is revealed to the table, the land enters tapped
// through the CR 614 pipeline without counting as a land play
// (CR 305.4), and the cards above it go to the bottom in an order drawn
// from the game's seeded RNG. A library with no land is revealed whole
// and goes to the bottom, which is the printed outcome.
//
// Crew is CrewCost / CrewEffect, the same as every other Vehicle.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "27a7610e-acbe-4a2b-9f61-81b383eb20a5",
		Name:            "The Regalia",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []ActivatedAbility{{
			Label:  "Crew 1",
			Cost:   CrewCost(1),
			Effect: CrewEffect("The Regalia"),
		}},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("The Regalia — reveal until a land card and put it onto the battlefield tapped",
				Do(RevealUntilThenPutOntoBattlefield{
					Match:  game.Card.IsLand,
					Tapped: true,
					Reason: "The Regalia — revealed until a land card",
				})),
		},
	})
}
