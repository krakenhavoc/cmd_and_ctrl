package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bewilder — Instant {2}{U} (Time Spiral, #5596):
//
//	"Target creature gets -3/-0 until end of turn.
//	 Draw a card."
//
// Fleeting Distraction at -3/-0: the shrink through
// boostTheTargetUntilEOT, then the draw. A target that became illegal
// counters the whole spell, draw included (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7cc945ed-d0aa-4b1a-a94d-656e03b6b5d3",
		Name:         "Bewilder",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(HarmUntilEOT(0, -3, 0)), Draws: 1},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := boostTheTargetUntilEOT(-3, 0, "Bewilder — -3/-0")(item, ctx); err != nil {
				return err
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
