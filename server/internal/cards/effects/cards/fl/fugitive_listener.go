package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fugitive Listener — Creature — El-Aurian Scout {2}{B}, 3/1:
//
//	"When this creature enters, each opponent discards a card."
//
// Burglar Rat's enters trigger on a bigger body: one discard prompt per
// opponent, each over that opponent's own hand, and an opponent with an
// empty hand is skipped (eachOpponentDiscardsOne).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e6939d1e-6ff8-4559-8526-0e4d0e749e20",
		Name:         "Fugitive Listener",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Fugitive Listener — each opponent discards a card", effects.EachOpponentDiscardsOne),
		},
	})
}
