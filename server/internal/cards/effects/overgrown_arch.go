package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Overgrown Arch — Creature — Plant Wall {1}{G}, 0/4 (EDHREC rank
// 20238):
//
//	"Defender
//	 {T}: You gain 1 life.
//	 {2}, Sacrifice this creature: Learn. (You may reveal a Lesson card
//	 you own from outside the game and put it into your hand, or discard
//	 a card to draw a card.)"
//
// The sacrifice is a cost (CR 602.2b), so the Arch is already in the
// graveyard when learn resolves; learn reads nothing off it
// (CR 701.48a, learn.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "802f8bcd-9219-48cd-a4c5-d9ea9ef60c9e",
		Name:            "Overgrown Arch",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: You gain 1 life.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    TapCost(),
				Effect:  gainOneLifeEffect,
			},
			{
				Label:   "{2}, Sacrifice this creature: Learn.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
				Cost:    Plus(ManaCost("{2}"), game.AbilityCost{SacrificeSelf: true}),
				Effect:  Do(Learn{}),
			},
		},
	})
}
