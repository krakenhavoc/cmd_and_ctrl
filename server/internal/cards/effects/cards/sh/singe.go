package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Singe — Instant {R} (Planeshift, S82 OLD #5615):
//
//	"Singe deals 1 damage to target creature. That creature becomes black until
//	 end of turn."
//
// The damage, then a layer-5 colour set (CR 105.3, 613.1e) lasting until the
// cleanup step (CR 514.2): the creature is black and no other colour. Both
// happen to the target if it is still legal as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e7fbdebf-2275-4bf6-b364-6a1849765f94",
		Name:         "Singe",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 1)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.OldB640DamageThenBecomesBlack(1, "Singe"),
	})
}
