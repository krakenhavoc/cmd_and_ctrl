package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Overprotect — Instant {1}{G} (Bloomburrow, #4467):
//
//	"Target creature you control gets +3/+3 and gains trample, hexproof,
//	 and indestructible until end of turn."
//
// One target, a pump and three keywords until end of turn on the
// still-legal target (pumpAndGrantFirstTarget, Adamant Will's body); a
// target that left in response, or one you no longer control, counters
// the spell (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c8c3a293-ae6d-4c52-be2d-2d4b6a8edb66",
		Name:         "Overprotect",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		Purpose:      ForTargets(BenefitUntilEOT(0, 3, 3, "trample", "hexproof", "indestructible")),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return pumpAndGrantFirstTarget(3, 3, "Overprotect — +3/+3, trample, hexproof, and indestructible",
				"trample", "hexproof", "indestructible")(ctx.Game, item)
		},
	})
}
