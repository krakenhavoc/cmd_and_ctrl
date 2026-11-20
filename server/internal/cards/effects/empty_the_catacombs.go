package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "49e0af9d-eb76-4ad6-b60a-430ae353728c",
		Name:         "Empty the Catacombs",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return eachPlayerReturnsCreatureCardsToHand(ctx)
		},
	})
}
