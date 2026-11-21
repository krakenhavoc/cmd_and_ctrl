package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Witchstalker Frenzy — Instant {3}{R} (Wilds of Eldraine, #4471):
//
//	"This spell costs {1} less to cast for each creature that attacked
//	 this turn.
//	 Witchstalker Frenzy deals 5 damage to target creature."
//
// Rowdy Research's discount (creaturesThatAttackedThisTurn): distinct
// creatures declared as attackers this turn, anyone's, through the one
// pricer (CR 601.2f). It reduces generic mana only, so the spell never
// costs less than {R}. Then Direct Hit's 5 damage to target creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f46befcb-6a7f-4f5e-8634-393cbf4b3308",
		Name:         "Witchstalker Frenzy",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 5)),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.CountCreaturesThatAttackedThisTurn,
				"This spell costs {1} less to cast for each creature that attacked this turn"),
		},
		OnResolve: effects.DamageToFirstTargetStep(5),
	})
}
