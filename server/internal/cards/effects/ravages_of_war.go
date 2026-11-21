package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "265236d1-d96c-4424-835e-a0fe91c41170",
		Name:         "Ravages of War",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: Land()}.Apply(ctx)
		},
	})
}
