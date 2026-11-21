package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thundering Broodwagon — Artifact — Vehicle {2}{B}{B}{G}{G}, 6/5
// (Aetherdrift, #4464):
//
//	"Reach, menace
//	 When this Vehicle enters, destroy target nonland permanent an
//	 opponent controls with mana value 4 or less.
//	 Crew 3
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Avalanche Riders' enters-destroy with the target narrowed to an
// opponent's nonland permanent of mana value 4 or less, read through the
// layers at announce and again at resolution (CR 608.2b). The Vehicle
// need not be crewed for the trigger, which watches it entering, not
// becoming a creature. The keywords are the Vehicle's whether or not it
// is crewed. Crew is the engine's (CR 702.122a); cycling is the hand
// ability (CR 702.29a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "817464d8-1e2d-4e12-96f8-5c7ec9b3e681",
		Name:            "Thundering Broodwagon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "menace"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Thundering Broodwagon — destroy target nonland permanent an opponent controls with mana value 4 or less",
					effects.DestroyFirstLegalCardTarget),
				effects.TargetPermanent("target nonland permanent an opponent controls with mana value 4 or less",
					effects.Nonland(), effects.OpponentControls(), effects.ManaValueLE(4))),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Crew 3",
				Cost:    effects.CrewCost(3),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  effects.CrewEffect("Thundering Broodwagon"),
			},
			effects.Cycling("{2}"),
		},
	})
}
