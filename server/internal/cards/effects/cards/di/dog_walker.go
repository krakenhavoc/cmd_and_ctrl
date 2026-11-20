package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dog Walker — Creature — Human Citizen {R}{W}, 3/1 (Murders at Karlov
// Manor, #4469):
//
//	"Vigilance
//	 Disguise {R/W}{R/W} (You may cast this card face down for {3} as a
//	 2/2 creature with ward {2}. Turn it face up any time for its disguise
//	 cost.)
//	 When this creature is turned face up, create two tapped 1/1 white Dog
//	 creature tokens."
//
// The face-up trigger (CR 708.8) makes the Dogs with the creation-time
// "tapped" (CreateTokenAdvanced, Token(...).EntersTapped()). Casting it
// face up never makes Dogs: the trigger watches only the turn face up.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "10d24900-cd40-4c98-bb1e-b9e66113f703",
		Name:             "Dog Walker",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"vigilance"},
		AlternativeCosts: []game.AlternativeCost{effects.Disguise("{R/W}{R/W}")},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.WhenThisIsTurnedFaceUp("Dog Walker — create two tapped 1/1 white Dog creature tokens",
					effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.TokenCard("1/1 white Dog")).EntersTapped(), N: 2})),
				game.Purpose{Tokens: 2}),
		},
	})
}
