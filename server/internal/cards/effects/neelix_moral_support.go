package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Neelix, Moral Support — {1}{G} Legendary Creature — Talaxian Advisor,
// 1/1 (Star Trek, #4973):
//
//	"When Neelix enters, create a Food token.
//	 Whenever you gain life, put a +1/+1 counter on Neelix."
//
// The Food is the shared printed template. The second ability is Ajani's
// Pridemate's: one counter per life-gain event, not per point, and a
// Neelix that has left gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7fba522b-a292-4d1f-9999-f787ba2d5fd1",
		Name:         "Neelix, Moral Support",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Neelix, Moral Support — create a Food token", Do(CreateToken{Template: FoodToken(), N: 1})),
			WheneverYouGainLife("Neelix, Moral Support — put a +1/+1 counter on Neelix", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
