package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Morbid Hunger — Sorcery {4}{B}{B} (Odyssey, #5613):
//
//	"Morbid Hunger deals 3 damage to any target. You gain 3 life.
//	 Flashback {7}{B}{B} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Lightning Helix's two sentences as steps: the life gain waits for any
// prompt the damage asks (CR 608.2c). With its only target gone by
// resolution the spell does nothing, the life gain included (CR
// 608.2b).
//
// Flashback is the shared constructor: the graveyard cast, the price,
// and CR 702.34a's exile as the spell leaves the stack (resolved,
// countered, or with its target gone).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "6e7e9685-ebe9-4db2-b52c-f6ce95e37e56",
		Name:             "Morbid Hunger",
		Completeness:     CompletenessFull,
		Targets:          TargetAny(),
		Purpose:          game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 3)), LifeGain: 3, Answers: game.AnswerRemove},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{7}{B}{B}")},
		OnResolve:        ResolveSteps(damageTheClauseTarget(0, 3), GainLife{Amount: 3}),
	})
}
