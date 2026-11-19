package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gift of Strength — Instant {1}{G} (Ravnica: Clue Edition, #5929):
//
//	"Target creature gets +3/+3 and gains reach until end of turn."
//
// Dive Down's body: the boost and the keyword land on the still-legal
// target (pumpAndGrantFirstTarget).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fe971240-268a-47d4-a14e-3d6352656091",
		Name:         "Gift of Strength",
		Purpose:      ForTargets(BenefitUntilEOT(0, 3, 3, "reach")),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return pumpAndGrantFirstTarget(3, 3, "Gift of Strength — +3/+3 and reach", "reach")(ctx.Game, item)
		},
	})
}
