package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ravages of War — Sorcery {3}{W} (Fallout):
//
//	"Destroy all lands."
//
// Armageddon, reprinted: every land, the caster's included, destroyed (not
// exiled), so an indestructible land survives and the rest go to their
// owners' graveyards. See armageddon.go for the reasoning.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "265236d1-d96c-4424-835e-a0fe91c41170",
		Name:         "Ravages of War",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.Land()}.Apply(ctx)
		},
	})
}
