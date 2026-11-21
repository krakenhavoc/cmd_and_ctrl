package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "4fd74a8c-a4e7-4988-ae02-9a73ab6ec89e",
		Name:         "Pictures of Spider-Man",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Pictures of Spider-Man — look at the top five cards, reveal up to two creature cards and put them into your hand",
				effects.LookAtTopThenMayTakeToHand(5, effects.Creature(), 2,
					"Pictures of Spider-Man — you may reveal up to two creature cards and put them into your hand")),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice this artifact: Create a Treasure token.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeThis()),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1}),
		}},
	})
}
