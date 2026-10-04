package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Birthday Escape — Sorcery {U}:
//
//	"Draw a card. The Ring tempts you."
//
// In printed order (CR 608.2c). A draw never pauses, so the tempt can
// follow it directly.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4f958678-2a32-4a30-8fe7-47df9ada6b2f",
		Name:         "Birthday Escape",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
				return err
			}
			return TheRingTemptsYou{}.Apply(ctx)
		},
	})
}
