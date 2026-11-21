package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "38d9401e-2840-455d-8478-8c976588d4ab",
		Name:         "Off Balance",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.NemTargetRestrictedThisTurn(game.CantAttackOrBlock, "Off Balance — can't attack or block this turn")(ctx.Game, item)
		},
	})
}
