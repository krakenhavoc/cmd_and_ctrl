package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inside Out — Instant {1}{U/R}:
//
//	"Switch target creature's power and toughness until end of turn.
//	 Draw a card."
//
// The switch is layer 7d (CR 613.4d, #3996): it applies after every
// set, pump and counter, whatever their timestamps. With the target gone the spell
// doesn't resolve (CR 608.2b), so no card is drawn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "623d9db1-a3af-4a1c-9643-975de364e073",
		Name:         "Inside Out",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := switchEachTargetPT(ctx, "Inside Out — switch power and toughness"); err != nil {
				return err
			}
			return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}
