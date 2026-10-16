package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Twisted Image — Instant {U}:
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
		OracleID:     "4a40cee1-fc09-4f97-aef8-440778d8c0ca",
		Name:         "Twisted Image",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := switchEachTargetPT(ctx, "Twisted Image — switch power and toughness"); err != nil {
				return err
			}
			return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}
