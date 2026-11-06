package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Misleading Motes — Instant {3}{U} (Wilds of Eldraine, #4471):
//
//	"Target creature's owner puts it on their choice of the top or bottom
//	 of their library."
//
// Run Behind without the discount: the owner, not the caster, picks the
// end of the library (ownerTucksFirstTargetTopOrBottom), so the prompt is
// addressed to that player. A creature that is gone or no longer a legal
// target as the spell resolves does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "09eec5a2-7835-4b24-9dd1-594612ee9152",
		Name:         "Misleading Motes",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ownerTucksFirstTargetTopOrBottom(ctx, "Misleading Motes — put it on your choice of the top or bottom of your library")
		},
	})
}
