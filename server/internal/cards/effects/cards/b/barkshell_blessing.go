package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Barkshell Blessing — Instant {G/W} (Shadowmoor, S82 older sets):
//
//	"Target creature gets +2/+2 until end of turn.
//	 Conspire (As you cast this spell, you may tap two untapped creatures you
//	 control that share a color with it. When you do, copy it and you may choose a new target for the copy.)"
//
// Mutagenic Growth's pump.
// Conspire is the engine's optional cost and cast trigger (conspire.go, CR
// 702.78), so the card declares the cost only. The conspire copy is a copy of the spell and may choose a new target (CR 707.10c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "baf3d4d9-0590-4d94-8f4e-1e726194b920",
		Name:          "Barkshell Blessing",
		Completeness:  effects.CompletenessFull,
		Purpose:       effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		Targets:       effects.TargetCreature("target creature"),
		OptionalCosts: []game.AdditionalCost{effects.Conspire()},
		OnResolve:     effects.BoostTheTargetUntilEOT(2, 2, "Barkshell Blessing — +2/+2"),
	})
}
