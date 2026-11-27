package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wandering Stream — Sorcery {2}{G} (Invasion, #5616):
//
//	"Domain — You gain 2 life for each basic land type among lands you
//	 control."
//
// The count is read once, as the spell resolves (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1f592b83-7838-4466-8bcb-093ac7346c55",
		Name:         "Wandering Stream",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GainLife{Amount: 2 * effects.DomainCount(ctx.Game, ctx.Controller())}.Apply(ctx)
		},
	})
}
