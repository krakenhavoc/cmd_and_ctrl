package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magitek Armor — Artifact — Vehicle {3}{W}, 4/4 (Final Fantasy, #4462):
//
//	"When this Vehicle enters, create a 1/1 colorless Hero creature token.
//	 Crew 1"
//
// The Hero is the token job select makes, here on its own; it can crew
// the Armor straight away, since crewing taps it and needs no haste.
// Crew 1 is the ordinary CR 702.122 crew ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0e9d8e12-b9da-4452-a758-5caffd8a752e",
		Name:         "Magitek Armor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Magitek Armor — create a 1/1 colorless Hero creature token",
				Do(CreateToken{Template: TokenCard(jobSelectHeroToken), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Magitek Armor"),
		}},
	})
}
