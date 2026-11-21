package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ashe, Princess of Dalmasca — Legendary Creature — Human Rebel Noble
// {2}{W}, 3/2 (Final Fantasy, #4462):
//
//	"Whenever Ashe attacks, look at the top five cards of your library.
//	 You may reveal an artifact card from among them and put it into
//	 your hand. Put the rest on the bottom of your library in a random
//	 order."
//
// Horn of the Mark's sentence (LookAtTopThenMayTakeToHand): only Ashe's
// controller sees the five, the artifact taken is revealed, and the rest
// go to the bottom in an order drawn from the game's seeded RNG.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "db75577d-1001-4088-999f-40684218a07f",
		Name:         "Ashe, Princess of Dalmasca",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Ashe, Princess of Dalmasca — look at the top five cards of your library",
				effects.LookAtTopThenMayTakeToHand(5, effects.Artifact(), 1,
					"Ashe — you may reveal an artifact card from among them and put it into your hand")),
		},
	})
}
