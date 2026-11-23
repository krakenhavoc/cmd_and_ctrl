package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scorching Shot — Sorcery {R}{R} (Outlaws of Thunder Junction, #4468):
//
//	"Scorching Shot deals 5 damage to target creature."
//
// Direct Hit's body (the timing is the card type's). A target that is
// gone or no longer a creature as it resolves takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e13113ba-1923-493c-a48d-ca320b7adbde",
		Name:         "Scorching Shot",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 5)),
		OnResolve:    effects.DamageToFirstTargetStep(5),
	})
}
