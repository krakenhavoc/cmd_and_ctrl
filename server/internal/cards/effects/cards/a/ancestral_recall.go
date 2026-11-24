package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ancestral Recall — Instant {U}: "Target player draws three cards."
// (Alpha, #5948). One of the Power Nine Oracle of the Alpha conjures
// (ADR 0099 amendment 2026-11-24).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "550c74d4-1fcb-406a-b02a-639a760a4380",
		Name:         "Ancestral Recall",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve:    effects.ResolveSteps(effects.Instruction(targetPlayerDrawsThree)),
	})
}

// targetPlayerDrawsThree is "target player draws three cards", the
// player re-checked at resolution (CR 608.2b).
func targetPlayerDrawsThree(ctx *effects.Context) error {
	who, ok := ctx.ClauseTarget(0)
	if !ok || who.Kind != game.TargetPlayer {
		return nil
	}
	return effects.DrawCards{Player: who.ID, N: 3}.Apply(ctx)
}
