package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Burglar Rat — Creature — Rat {1}{B}, 1/1 (EDHREC rank 2601):
//
//	"When this creature enters, each opponent discards a card."
//
// The two-mana Rat that makes the table discard. An ETB trigger
// whose body is The Eldest Reborn's second chapter: one discard
// prompt per opponent, each addressed to that opponent and offering
// only their own hand, so nobody picks for anyone else
// (b24EachOpponentDiscardsOne). An opponent with no cards in hand is
// skipped. Each discard fires the discard payoffs (Megrim) as
// printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2f807301-37df-4724-871a-08e3512b07b3",
		Name:         "Burglar Rat",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Burglar Rat — each opponent discards a card", b24EachOpponentDiscardsOne),
		},
	})
}
