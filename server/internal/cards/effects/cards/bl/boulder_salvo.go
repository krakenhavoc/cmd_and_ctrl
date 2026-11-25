package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boulder Salvo — Sorcery {4}{R} (Oath of the Gatewatch, #5545):
//
//	"Surge {1}{R} (You may cast this spell for its surge cost if you or a
//	 teammate has cast another spell this turn.)
//	 Boulder Salvo deals 4 damage to target creature."
//
// Surge is the shared gated alternative cost (CR 702.117a); the spell is the
// same either way. The damage is skipped if the target is no longer legal as
// it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "16df20e2-5228-4543-84b7-74a3170d9d31",
		Name:             "Boulder Salvo",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Surge("{1}{R}")},
		Targets:          effects.TargetCreature("target creature"),
		Purpose: game.Purpose{
			Answers: game.AnswerRemove,
			Targets: game.ForTargets(effects.DamageToTarget(0, 4)),
		},
		OnResolve: effects.DamageToFirstTargetStep(4),
	})
}
