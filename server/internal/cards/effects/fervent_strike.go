package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fervent Strike — Instant {R} (Dominaria, #5530):
//
//	"Target creature gets +1/+0 and gains first strike and haste until end of turn."
//
// The target gets +1/+0 and gains first strike and haste until end of
// turn: one target, read as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "35efb1e6-6be2-4a38-969a-199524d48e02",
		Name:         "Fervent Strike",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 0, "first strike", "haste")),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return firstTargetGainsUntilEndOfTurn(1, "Fervent Strike — +1/+0, first strike and haste", "first strike", "haste")(ctx.Game, item)
		},
	})
}
