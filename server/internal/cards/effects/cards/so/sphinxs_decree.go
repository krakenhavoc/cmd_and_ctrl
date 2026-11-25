package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sphinx's Decree — Sorcery {1}{W}:
//
//	"Each opponent can't cast instant or sorcery spells during that player's
//	 next turn."
//
// One cast ban per opponent still in the game, over instants and sorceries,
// that applies only during that opponent's next turn (CastBanRule.OnTurnBegun)
// and ends with it: casts before that turn, and on any other player's turn,
// are untouched.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7d615b37-8dc6-4142-bf9a-9815dd99025a",
		Name:         "Sphinx's Decree",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB220NextTurnInstantSorceryBan(ctx, "Sphinx's Decree — can't cast instant or sorcery spells during your next turn")
		},
	})
}
