package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Woodcutter's Grit — Instant {2}{G} (Eldritch Moon, #5541):
//
//	"Target creature you control gets +3/+3 and gains hexproof until end
//	 of turn."
//
// Blossoming Defense at +3/+3 (pumpAndGrantFirstTarget).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7756d2b9-7b10-4d05-8ad3-8d8f39c1eb23",
		Name:         "Woodcutter's Grit",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 3, 3, "hexproof")),
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return pumpAndGrantFirstTarget(3, 3, "Woodcutter's Grit — +3/+3 and hexproof", "hexproof")(ctx.Game, item)
		},
	})
}
