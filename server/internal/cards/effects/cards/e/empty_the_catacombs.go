package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Empty the Catacombs — Sorcery {3}{B} (Ravnica: City of Guilds, #5600):
//
//	"Each player returns all creature cards from their graveyard to their
//	 hand."
//
// Underworld Cerberus's dies sweep as a spell: every creature card in
// every graveyard returns to its owner's hand at once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49e0af9d-eb76-4ad6-b60a-430ae353728c",
		Name:         "Empty the Catacombs",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.EachPlayerReturnsCreatureCardsToHand(ctx)
		},
	})
}
