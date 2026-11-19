package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mighty Leap — Instant {1}{W} (Ravnica: Clue Edition, #5929):
//
//	"Target creature gets +2/+2 and gains flying until end of turn."
//
// Dive Down's body: the boost and the keyword land on the still-legal
// target (pumpAndGrantFirstTarget).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "593caa15-e1d1-477f-bbb4-124d7a7b81f3",
		Name:         "Mighty Leap",
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2, "flying")),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return pumpAndGrantFirstTarget(2, 2, "Mighty Leap — +2/+2 and flying", "flying")(ctx.Game, item)
		},
	})
}
