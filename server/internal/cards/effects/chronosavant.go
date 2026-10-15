package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chronosavant — Creature — Giant {5}{W}, 5/5:
//
//	"{1}{W}: Return this card from your graveyard to the battlefield
//	 tapped. You skip your next turn."
//
// An ability that works from the graveyard (CR 113.6, Zones) with the
// Reassembling Skeleton return, then the skip (CR 614.10, ADR 0188
// Decision 1, #3565). The skip is its own sentence: it is written even
// when the card left the graveyard in response and nothing came back.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "969c9864-f78b-46b2-b35b-c9adaf355c2d",
		Name:         "Chronosavant",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}: Return this card from your graveyard to the battlefield tapped. You skip your next turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue, SkipsTurns: 1},
			Cost:    ManaCost("{1}{W}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := returnThisFromGraveyardTapped(g, item); err != nil {
					return err
				}
				return SkipYourNextTurns{}.Apply(NewContext(g, item))
			},
		}},
	})
}
