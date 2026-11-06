package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Give In to Violence — Instant {1}{B} (Duskmourn: House of Horror,
// #4466):
//
//	"Target creature gets +2/+2 and gains lifelink until end of turn."
//
// Adamant Will's body with lifelink in place of indestructible
// (pumpAndGrantFirstTarget): a target gone by resolution gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2bf4811f-3462-4542-acb0-101e772bfec9",
		Name:         "Give In to Violence",
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2, "lifelink")),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return pumpAndGrantFirstTarget(2, 2, "Give In to Violence — +2/+2 and lifelink", "lifelink")(ctx.Game, item)
		},
	})
}
