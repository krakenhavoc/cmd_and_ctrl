package effects

// Ragefire — Sorcery {1}{R} (Ravnica Allegiance, #5521):
//
//	"Ragefire deals 3 damage to target creature."
//
// Repulsor Rays's sentence: 3 damage to the creature
// (damageToFirstTarget).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "91b8c0bc-2599-4423-a23f-5dc95b23cf25",
		Name:         "Ragefire",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(DamageToTarget(0, 3)),
		OnResolve:    damageToFirstTarget(3),
	})
}
