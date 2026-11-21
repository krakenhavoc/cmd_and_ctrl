package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Magnificent End — Instant {4}{W} (The Hobbit):
//
//	"This spell costs {3} less to cast if it targets a tapped creature.
//	 Magnificent End deals 5 damage to target creature."
//
// #4454, S76. The reduction reads the announced target (CR 601.2f, after
// 601.2c): CostsLessIfItTargets, as Price of Fame does. Whether the
// creature is tapped is judged when the cost is determined; the damage
// is dealt to whatever the target is on resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "82bdf9cb-61fe-4eaf-8762-c32002403d96",
		Name:         "Magnificent End",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessIfItTargets(3, "This spell costs {3} less to cast if it targets a tapped creature.", effects.HobB3TappedCreature()),
		},
		Purpose:   effects.ForTargets(effects.DamageToTarget(0, 5)),
		OnResolve: effects.DamageToFirstTargetStep(5),
	})
}
