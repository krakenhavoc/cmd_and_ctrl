package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Curtains' Call — Instant {5}{B}:
//
//	"Undaunted (This spell costs {1} less to cast for each opponent.)
//	 Destroy two target creatures."
//
// #746: undaunted (CR 702.125a) is a self cost modifier counting the
// caster's opponents still in the game. A target that became illegal
// in response is skipped, and the other is still destroyed.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cfb1b64b-bf78-4015-ac29-cdcdf2adaa02",
		Name:         "Curtains' Call",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("two target creatures").WithCount(2, 2),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.OpponentsOfCaster(), "Undaunted"),
		},
		OnResolve: effects.DestroyEachLegalTarget,
	})
}
