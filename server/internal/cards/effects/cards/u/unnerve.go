package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unnerve — Sorcery {3}{B} (Conspiracy: Take the Crown, S82 older sets
// batch 271, #5540):
//
//	"Each opponent discards two cards."
//
// Nothing targets. Each opponent's discard is their own choice of which
// two cards (EachOpponentDiscards), and an opponent with fewer than
// two cards discards what they have (CR 701.9b); a player who has left the
// game is not asked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8386f43f-8106-415c-a9a6-7fd3f5b060f3",
		Name:         "Unnerve",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			effects.EachOpponentDiscards(ctx, 2)
			return nil
		},
	})
}
