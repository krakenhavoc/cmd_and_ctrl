package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Off Balance — Instant {W}
// (Nemesis, #5618):
//
//	"Target creature can't attack or block this turn."
//
// Change of Heart's restriction with both halves
// (game.CantAttackOrBlock), through nemTargetRestrictedThisTurn. It
// ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "38d9401e-2840-455d-8478-8c976588d4ab",
		Name:         "Off Balance",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(HarmUntilEOT(0, 0, 0)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return nemTargetRestrictedThisTurn(game.CantAttackOrBlock, "Off Balance — can't attack or block this turn")(ctx.Game, item)
		},
	})
}
