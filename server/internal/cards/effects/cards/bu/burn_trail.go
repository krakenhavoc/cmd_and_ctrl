package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Burn Trail — Sorcery {3}{R} (Shadowmoor, S82 older sets):
//
//	"Burn Trail deals 3 damage to any target.
//	 Conspire (As you cast this spell, you may tap two untapped creatures you
//	 control that share a color with it. When you do, copy it and you may choose a new target for the copy.)"
//
// Lightning Bolt's body for 3 at sorcery speed with conspire.
// Conspire is the engine's optional cost and cast trigger (conspire.go, CR
// 702.78), so the card declares the cost only. The conspire copy is a copy of the spell and may choose a new target (CR 707.10c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "566b4dcc-b7d7-4443-b2d0-954f1fc33a11",
		Name:          "Burn Trail",
		Completeness:  effects.CompletenessFull,
		Purpose:       effects.ForTargets(effects.DamageToTarget(0, 3)),
		Targets:       effects.TargetAny(),
		OptionalCosts: []game.AdditionalCost{effects.Conspire()},
		OnResolve:     effects.DamageToFirstTargetStep(3),
	})
}
