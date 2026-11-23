package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mortality Spear — Instant {2}{B}{G}:
//
//	"This spell costs {2} less to cast if you gained life this turn.
//	 Destroy target nonland permanent."
//
// #746: a conditional self cost modifier, reading the turn tally's
// life gained for the caster.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "544acde0-850d-4c4f-8389-e22930d87345",
		Name:         "Mortality Spear",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(2, "This spell costs {2} less to cast if you gained life this turn.", effects.YouGainedLifeThisTurn()),
		},
		OnResolve: effects.DestroyTheTargetPermanent,
	})
}
