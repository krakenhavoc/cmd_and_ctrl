package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Feat of Resistance — Instant {1}{W} (Khans of Tarkir; Jumpstart):
//
//	"Put a +1/+1 counter on target creature you control. It gains
//	 protection from the color of your choice until end of turn."
//
// The counter first, then Center Soul's colour choice: the colour is
// asked as the spell resolves, after the counter has landed, and the
// creature gains protection from it until end of turn. A target that
// left in response is asked nothing and gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	protection := protectionFromAChosenColorForTheTarget("Feat of Resistance")
	Register(Spec{
		OracleID:     "2038f505-7b03-4c63-bc44-c5139729dd05",
		Name:         "Feat of Resistance",
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "protection")),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := PlusOneCounterOnTheTarget(ctx.Game, item); err != nil {
				return err
			}
			return protection(item, ctx)
		},
	})
}
