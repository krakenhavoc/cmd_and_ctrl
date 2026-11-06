package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Adamant Will — Instant {1}{W} (Foundations, #4465):
//
//	"Target creature gets +2/+2 and gains indestructible until end of
//	 turn."
//
// Hamato Guardian Stance's shape without the scry: the pump and the
// keyword land on the still-legal target (pumpAndGrantFirstTarget).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "467e22c3-6107-40ff-afc7-5960710c970b",
		Name:         "Adamant Will",
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2, "indestructible")),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return pumpAndGrantFirstTarget(2, 2, "Adamant Will — +2/+2 and indestructible", "indestructible")(ctx.Game, item)
		},
	})
}
