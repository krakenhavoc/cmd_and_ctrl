package effects

// Command the Storm — Instant {4}{R} (Guilds of Ravnica, #5525):
//
//	"Command the Storm deals 5 damage to target creature."
//
// Concentrated Fire's body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1dc816b5-8918-4a0c-bb10-74abe579a98c",
		Name:         "Command the Storm",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(DamageToTarget(0, 5)),
		OnResolve:    damageToFirstTarget(5),
	})
}
