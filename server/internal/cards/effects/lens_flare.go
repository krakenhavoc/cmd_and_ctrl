package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "e39bcfb8-cc45-42ab-a4b6-003f44d6b578",
		Name:         "Lens Flare",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(DamageToTarget(0, 5)),
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for artifacts", Artifact()),
		},
		Targets:   TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
		OnResolve: damageToFirstTarget(5),
	})
}
