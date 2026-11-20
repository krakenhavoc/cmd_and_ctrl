package effects

// Arrows of Justice — Instant {2}{R/W} (Gatecrash, #5565):
//
//	"Arrows of Justice deals 4 damage to target attacking or blocking
//	 creature."
//
// Divine Arrow's spell: 4 damage to the one target, which must still be
// attacking or blocking as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "46e311ef-c86e-4aad-8b42-ef0d68a90c2e",
		Name:         "Arrows of Justice",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(DamageToTarget(0, 4)),
		Targets:      TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
		OnResolve:    damageToFirstTarget(4),
	})
}
