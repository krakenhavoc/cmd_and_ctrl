package effects

// Electrify — Instant {3}{R}
// (Amonkhet, #5535):
//
//	"Electrify deals 4 damage to target creature."
//
// Flame Slash's body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "da952213-9a3c-4e13-bc5a-55d1e5dbaa5b",
		Name:         "Electrify",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(DamageToTarget(0, 4)),
		OnResolve:    damageToFirstTarget(4),
	})
}
