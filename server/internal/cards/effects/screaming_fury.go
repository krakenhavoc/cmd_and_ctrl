package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Screaming Fury — Sorcery {2}{R} (Fifth Dawn, #5604):
//
//	"Target creature gets +5/+0 and gains haste until end of turn."
//
// Wild Ride's +N/+0 and haste, at five.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eaded717-e0e2-4def-be7e-6bfb137628c2",
		Name:         "Screaming Fury",
		Purpose:      ForTargets(BenefitUntilEOT(0, 5, 0, "haste")),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return firstTargetGainsUntilEndOfTurn(5, "Screaming Fury — +5/+0 and haste", "haste")(ctx.Game, item)
		},
	})
}
