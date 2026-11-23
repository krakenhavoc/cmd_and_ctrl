package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gang Up — Instant {X}{B}:
//
//	"Assist (Another player can pay up to {X} of this spell's cost. You
//	 choose the value of X.)
//	 Destroy target creature with power X or less."
//
// Killing Glare's card, with the same announced-X power bound
// (ADR 0109 §9). X is chosen before the target (CR 601.2b-c) and the
// helper's share is capped at the X in the total cost (ADR 0150 §4),
// so the assist needs nothing beyond the keyword.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dea311c3-9be3-4059-a206-5fbbdca4d6c9",
		Name:            "Gang Up",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordAssist},
		Targets:         effects.TargetCreature("target creature with power X or less").WithPowerAtMostX(),
		OnResolve:       effects.DestroyTheTargetPermanent,
	})
}
