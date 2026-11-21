package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Squash — Instant {4}{R} (Kaldheim, #5505):
//
//	"This spell costs {3} less to cast if you control a Giant.
//	 Squash deals 6 damage to target creature or planeswalker."
//
// The reduction is Academy Journeymage's self cost modifier
// (CostsLess with YouControlA) read for a Giant, any permanent with the
// creature type; the damage is Lightning Bolt's body
// (damageToFirstTarget) at 6.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "54bd7b7c-4052-437d-b2e4-311339e8e16e",
		Name:         "Squash",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 6)),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(3, "This spell costs {3} less to cast if you control a Giant.", effects.YouControlA(effects.HasSubtype("Giant"))),
		},
		OnResolve: effects.DamageToFirstTargetStep(6),
	})
}
