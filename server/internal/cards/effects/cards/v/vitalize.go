package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vitalize — Instant {G} (EDHREC rank 3555):
//
//	"Untap all creatures you control."
//
// The one-mana mass untap. Every tapped creature the caster controls
// untaps; nobody else's does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd3226c6-bd47-4246-8ab3-01316c5d1809",
		Name:         "Vitalize",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.B34UntapAllCreaturesYouControl(ctx.Game, item)
		},
	})
}
