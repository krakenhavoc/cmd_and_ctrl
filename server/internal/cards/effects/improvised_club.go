package effects

// Improvised Club — Instant {1}{R} (The Hobbit Commander):
//
//	"As an additional cost to cast this spell, sacrifice an artifact or
//	 creature.
//	 Improvised Club deals 4 damage to any target."
//
// #4454, S76. Deadly Dispute's additional cost (paid with the spell
// already on the stack, CR 601.2h) and Lightning Bolt's damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "20c705e3-4141-41f8-ad73-bf953f0e1ddd",
		Name:           "Improvised Club",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("an artifact or creature", Or(Artifact(), Creature())),
		Targets:        TargetAny(),
		Purpose:        ForTargets(DamageToTarget(0, 4)),
		OnResolve:      damageToFirstTarget(4),
	})
}
