package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Burglar Rat — Creature — Rat {1}{B}, 1/1 (EDHREC rank 2601):
//
//	"When this creature enters, each opponent discards a card."
//
// The two-mana Rat that makes the table discard. An ETB trigger
// whose body is The Eldest Reborn's second chapter: one discard
// prompt per opponent, each addressed to that opponent and offering
// only their own hand, so nobody picks for anyone else
// (eachOpponentDiscardsOne). An opponent with no cards in hand is
// skipped. Each discard fires the discard payoffs (Megrim) as
// printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f807301-37df-4724-871a-08e3512b07b3",
		Name:         "Burglar Rat",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Burglar Rat — each opponent discards a card", effects.EachOpponentDiscardsOne),
		},
	})
}
