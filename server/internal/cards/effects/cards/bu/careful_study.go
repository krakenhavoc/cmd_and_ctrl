package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Careful Study — Sorcery {U} (EDHREC rank 2713):
//
//	"Draw two cards, then discard two cards."
//
// The blue Faithless Looting, and the same body: the draw resolves
// first and the discard choice is queued against the post-draw hand,
// so the two cards just drawn are legal discards exactly as in
// paper. The caster picks the discards in the hand prompt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "32e9aa23-fc0a-4b82-82f9-1659c304428c",
		Name:         "Careful Study",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.LootOne(ctx.Game, item, 2)
		},
	})
}
