package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Templar Knight — Creature — Human Knight {1}{W}, 3/1 (Assassin's Creed,
// #5948):
//
//	"Vigilance
//	 {W}, Tap five untapped attacking creatures you control named Templar
//	 Knight: Search your library for a legendary artifact card, put it onto the
//	 battlefield, then shuffle.
//	 A deck can have any number of cards named Templar Knight."
//
// The cost is {W} and the tapping of five untapped attacking creatures you
// control with the name (TapOthers with a count of five, chosen at activation;
// the Knight itself may be one of them, since the clause says "creatures", not
// "other"). It is not the {T} symbol (CR 302.6), so summoning sickness does not
// apply, and attacking Knights are still untapped because they have vigilance. A
// face-down creature has no name and is never one of them. The search names a
// quality, so it may find nothing (CR 701.23b), and shuffles either way. The deck
// clause is a deck-construction rule (CR 100.2a) that internal/deck reads off the
// oracle text; nothing in the game depends on it.
//
// No simplification.
func init() {
	const label = "five untapped attacking creatures you control named Templar Knight"
	Register(Spec{
		OracleID:        "f9453fe2-fadf-4cd4-8d2c-0eaa0e2d78d6",
		Name:            "Templar Knight",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Purpose:         game.Purpose{Tutors: 1},
		Activated: []ActivatedAbility{{
			Label:   "{W}, Tap five untapped attacking creatures you control named Templar Knight: Search your library for a legendary artifact card, put it onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost: Plus(ManaCost("{W}"), game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count: 5,
				Filter: TargetPermanent(label, Creature(), YouControl(),
					func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
						return c.AttackingTarget != uuid.Nil && !c.FaceDownIsPermanent() && c.Name == "Templar Knight"
					}),
				Label: label,
			}}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.IsArtifact() && c.IsLegendary() },
					Dest:      game.ZoneBattlefield,
					Limit:     1,
					Shuffle:   true,
					Reason:    "Templar Knight — a legendary artifact card",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
