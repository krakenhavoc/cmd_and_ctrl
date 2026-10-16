package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// About Face — Instant {R}:
//
//	"Switch target creature's power and toughness until end of turn."
//
// The switch is layer 7d (CR 613.4d, #3996): it applies after every
// set, pump and counter, whatever their timestamps.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8a810017-06ac-4824-bc2f-a7d303d58ad7",
		Name:         "About Face",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return switchEachTargetPT(ctx, "About Face — switch power and toughness")
		},
	})
}
