package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inspired Fire — {3}{R} Instant (Marvel Super Heroes Commander, #4455):
//
//	"This spell costs {1} less to cast for each card you've drawn this turn.
//	 Inspired Fire deals 4 damage to target creature."
//
// The reduction is a self cost modifier (Blasphemous Act's shape) reading the
// per-turn draw tally, so it applies to generic mana only and never makes
// the spell cost less than {R}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7f43b15b-b5d4-4ec6-bd5b-9cf56b2a6bde",
		Name:         "Inspired Fire",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(func(q game.CostQuery) int {
				if q.Game == nil {
					return 0
				}
				return q.Game.TurnTallyFor(q.Controller).CardsDrawn
			}, "This spell costs {1} less to cast for each card you've drawn this turn."),
		},
		Targets:   effects.TargetCreature("target creature"),
		Purpose:   effects.ForTargets(effects.DamageToTarget(0, 4)),
		OnResolve: effects.DamageToFirstTargetStep(4),
	})
}
