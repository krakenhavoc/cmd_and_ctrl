package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "82bdf9cb-61fe-4eaf-8762-c32002403d96",
		Name:         "Magnificent End",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfItTargets(3, "This spell costs {3} less to cast if it targets a tapped creature.", hobB3TappedCreature()),
		},
		Purpose:   ForTargets(DamageToTarget(0, 5)),
		OnResolve: damageToFirstTarget(5),
	})
}
