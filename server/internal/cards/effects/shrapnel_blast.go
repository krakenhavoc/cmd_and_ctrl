package effects

// Shrapnel Blast — Instant {1}{R} (Mirrodin, #5606):
//
//	"As an additional cost to cast this spell, sacrifice an artifact.
//	 Shrapnel Blast deals 5 damage to any target."
//
// Forge Armor's sacrifice, paid as the spell is cast (CR 601.2h), so it
// can't be cast without an artifact to sacrifice; then Thunderous
// Wrath's 5 damage to a target checked again as it resolves
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "d3c623f3-c7c9-47c0-8096-6dee3267baff",
		Name:           "Shrapnel Blast",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("an artifact", Artifact()),
		Targets:        TargetAny(),
		Purpose:        ForTargets(DamageToTarget(0, 5)),
		OnResolve:      damageToFirstTarget(5),
	})
}
