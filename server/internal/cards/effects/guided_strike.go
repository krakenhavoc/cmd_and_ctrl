package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Guided Strike — Instant {1}{W} (Weatherlight, #5629):
//
//	"Target creature gets +1/+0 and gains first strike until end of turn.
//	 Draw a card."
//
// Interjection's pump and grant at +1/+0, then the draw. With its only
// target gone the spell does not resolve, so there is no draw either
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a215e813-df7e-4f21-9718-eb264eb62813",
		Name:         "Guided Strike",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      game.Purpose{Draws: 1, Targets: game.ForTargets(BenefitUntilEOT(0, 1, 0, "first strike"))},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := socB27TargetGetsAndGains(1, 0, "first strike", "Guided Strike")(item, ctx); err != nil {
				return err
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
