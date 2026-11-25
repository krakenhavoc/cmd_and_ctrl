package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fork in the Road — Sorcery {1}{G} (Shadows over Innistrad, #5543):
//
//	"Search your library for up to two basic land cards and reveal them.
//	 Put one into your hand and the other into your graveyard. Then
//	 shuffle."
//
// Intuition's shape with the searcher choosing: the search reveals up to
// two basic lands and leaves them on top, then the searcher picks the one
// that goes to their hand, the other going to the graveyard, and the
// library is shuffled (effects.S82OldB284ForkInTheRoad). Finding only one
// basic land puts it in the hand; finding none is allowed (a search with a
// stated quality, CR 701.23b) and the library is still shuffled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3bc12623-3852-42f0-b458-5c67fe1b7a07",
		Name:         "Fork in the Road",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB284ForkInTheRoad(ctx)
		},
	})
}
