package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Delirium Skeins — Sorcery {2}{B} (Dissension, #5598):
//
//	"Each player discards three cards."
//
// Rankle's Prank's discard: every player, in APNAP order, picks three
// cards of their own; a player with fewer discards their hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7397036f-8114-47cc-b52f-c532a6845d16",
		Name:         "Delirium Skeins",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ctx.Game.EachPlayerDiscardsForEffect(uuid.Nil, game.DiscardPrompt{
				Source: item.SourceCardID, N: 3, Question: "Delirium Skeins — discard three cards",
			})
			return nil
		},
	})
}
