package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Comet Storm — Instant {X}{R}{R} (Worldwake; OLD b079, #6150):
//
//	"Multikicker {1} (You may pay an additional {1} any number of times
//	 as you cast this spell.)
//	 Choose any target, then choose another target for each time this
//	 spell was kicked. Comet Storm deals X damage to each of them."
//
// The clause is counted by the kicks (CountedByKicks, CR 601.2c): one
// target unkicked, one more for each payment of {1}, all different. Each
// target still legal as it resolves is dealt X damage at once (CR
// 608.2b). The multikicker cap of 20 is the engine's bound on one
// announcement, far above what a board can pay.
//
// No simplification.
func init() {
	targets := effects.CountedByKicks(effects.TargetAny())
	targets.Label = "any target, and another for each time this spell was kicked"
	effects.Register(effects.Spec{
		OracleID:      "9c2b262b-53c0-41a1-8f1a-0c336453d4bf",
		Name:          "Comet Storm",
		Completeness:  effects.CompletenessFull,
		XMatters:      true,
		OptionalCosts: []game.AdditionalCost{effects.Multikicker("{1}", 20)},
		Targets:       targets,
		Purpose:       effects.ForTargets(effects.DamageXToTarget(0)),
		OnResolve:     effects.DamageXToEachLegalTarget(),
	})
}
