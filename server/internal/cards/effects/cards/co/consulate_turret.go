package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Consulate Turret — Artifact {3}:
//
//	"{T}: You get {E} (an energy counter).
//	 {T}, Pay {E}{E}{E}: This artifact deals 2 damage to target player or
//	 planeswalker."
//
// ADR 0129 PR 1 (#1995).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "14d71432-c2e6-4c81-92ef-cf98e1fcf5f8",
		Name:         "Consulate Turret",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: You get {E}.",
				Cost:    effects.TapCost(),
				Purpose: game.Purpose{Answers: game.AnswerValue, Energy: 1},
				Effect:  effects.EbYouGetEnergy(1),
			},
			{
				Label:   "{T}, Pay {E}{E}{E}: This artifact deals 2 damage to target player or planeswalker.",
				Cost:    effects.Plus(effects.TapCost(), effects.PayEnergy(3)),
				Targets: effects.TargetPlayerOrPlaneswalker(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
				Effect:  effects.SourceDealsDamageToEachLegalTarget(2),
			},
		},
	})
}
