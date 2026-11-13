package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Journeyer's Kite — Artifact {2} (Champions of Kamigawa, #5603):
//
//	"{3}, {T}: Search your library for a basic land card, reveal it, put
//	 it into your hand, then shuffle."
//
// A repeatable land tutor to hand; the card found is revealed (CR
// 701.20a). It is not a creature, so it can be tapped the turn it
// arrives (CR 302.6 applies to creatures only).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "10aab5bc-5758-443b-ba4c-49f6c6d91262",
		Name:         "Journeyer's Kite",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}: Search your library for a basic land card, reveal it, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tutors: 1},
			Cost:    Plus(ManaCost("{3}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Journeyer's Kite — choose a basic land card to put into your hand",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
