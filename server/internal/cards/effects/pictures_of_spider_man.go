package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pictures of Spider-Man — Artifact {2}{G} (Marvel's Spider-Man):
//
//	"When this artifact enters, look at the top five cards of your
//	 library. You may reveal up to two creature cards from among them
//	 and put them into your hand. Put the rest on the bottom of your
//	 library in a random order.
//	 {1}, {T}, Sacrifice this artifact: Create a Treasure token."
//
// Horn of the Mark's dig (LookAtTopThenMayTakeToHand) with a ceiling of
// two; "up to two" is an optional pick of zero, one or two.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4fd74a8c-a4e7-4988-ae02-9a73ab6ec89e",
		Name:         "Pictures of Spider-Man",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Pictures of Spider-Man — look at the top five cards, reveal up to two creature cards and put them into your hand",
				LookAtTopThenMayTakeToHand(5, Creature(), 2,
					"Pictures of Spider-Man — you may reveal up to two creature cards and put them into your hand")),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice this artifact: Create a Treasure token.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}"), TapCost(), SacrificeThis()),
			Effect:  Do(CreateToken{Template: TreasureToken(), N: 1}),
		}},
	})
}
