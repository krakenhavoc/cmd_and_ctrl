package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lens Flare — Instant {4}{W} (Modern Horizons 2, #5501):
//
//	"Affinity for artifacts (This spell costs {1} less to cast for each
//	 artifact you control.)
//	 Lens Flare deals 5 damage to target attacking or blocking creature."
//
// Frogmite's affinity (the shared self cost modifier, CR 702.41a,
// generic mana only) on Osseous Exhale's five damage to a combatant.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e39bcfb8-cc45-42ab-a4b6-003f44d6b578",
		Name:         "Lens Flare",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 5)),
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for artifacts", effects.Artifact()),
		},
		Targets:   effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
		OnResolve: effects.DamageToFirstTargetStep(5),
	})
}
