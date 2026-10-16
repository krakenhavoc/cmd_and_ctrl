package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Transmutation — Instant — Arcane {1}{B}:
//
//	"Switch target creature's power and toughness until end of turn."
//
// The switch is layer 7d (CR 613.4d, #3996): it applies after every
// set, pump and counter, whatever their timestamps.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "72a75113-4ddf-460d-8d9b-03931c2788da",
		Name:         "Transmutation",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return switchEachTargetPT(ctx, "Transmutation — switch power and toughness")
		},
	})
}
