package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wildsize — Instant {2}{G} (Ravnica: Clue Edition):
//
//	"Target creature gets +2/+2 and gains trample until end of turn.
//	 Draw a card."
//
// Massive Might's body, then the controller draws. A target that has become
// illegal counters the whole spell, draw included (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e1f439c7-217f-4b20-a77d-b9014c5de8c0",
		Name:         "Wildsize",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Targets: game.ForTargets(BenefitUntilEOT(0, 2, 2, "trample"))},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := pumpAndGrantFirstTarget(2, 2, "Wildsize — +2/+2 and trample", "trample")(ctx.Game, item); err != nil {
				return err
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
