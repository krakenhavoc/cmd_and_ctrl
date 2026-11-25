package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thresher Lizard — {2}{R} Creature — Lizard 3/2 (Amonkhet, S82 old batch 246):
//
//	"This creature gets +1/+2 as long as you have one or fewer cards in hand."
//
// A layer 7c modification (CR 613.4c) read live from the controller's hand
// (SelfPumpWhileYouHaveAtMostCardsInHand).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "46a69558-fad1-4a04-9bb6-8f971bbbb038",
		Name:         "Thresher Lizard",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.SelfPumpWhileYouHaveAtMostCardsInHand(1, 1, 2)},
	})
}
