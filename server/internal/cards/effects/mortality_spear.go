package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mortality Spear — Instant {2}{B}{G}:
//
//	"This spell costs {2} less to cast if you gained life this turn.
//	 Destroy target nonland permanent."
//
// #746: a conditional self cost modifier, reading the turn tally's
// life gained for the caster.
func init() {
	Register(Spec{
		OracleID:     "544acde0-850d-4c4f-8389-e22930d87345",
		Name:         "Mortality Spear",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target nonland permanent", Nonland()),
		SelfCostModifiers: []game.CostModifier{
			CostsLess(2, "This spell costs {2} less to cast if you gained life this turn.", YouGainedLifeThisTurn()),
		},
		OnResolve: destroyTheTargetPermanent,
	})
}
