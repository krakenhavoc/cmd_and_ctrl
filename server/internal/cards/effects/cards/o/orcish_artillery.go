package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orcish Artillery — Creature — Orc Warrior {1}{R}{R}, 1/3 (Eighth Edition, S82
// OLD batch 611, #5607):
//
//	"{T}: This creature deals 2 damage to any target and 3 damage to you."
//
// Fireslinger's tap ability at 2 and 3 (TmtB6DamageAnyTargetAndYou): both are
// dealt at once, as one damage instance; if the target is gone the ability does
// nothing, the damage to you included (CR 608.2b). A tap ability, so summoning
// sickness applies (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c000811e-bde7-4840-a3bb-9714d5c977eb",
		Name:         "Orcish Artillery",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 2 damage to any target and 3 damage to you.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny(),
			Purpose: game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 2)), Answers: game.AnswerRemove},
			Effect:  effects.TmtB6DamageAnyTargetAndYou(2, 3),
		}},
	})
}
