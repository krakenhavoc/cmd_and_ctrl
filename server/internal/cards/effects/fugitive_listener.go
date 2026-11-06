package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "e6939d1e-6ff8-4559-8526-0e4d0e749e20",
		Name:         "Fugitive Listener",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Fugitive Listener — each opponent discards a card", eachOpponentDiscardsOne),
		},
	})
}
