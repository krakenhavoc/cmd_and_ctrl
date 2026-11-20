package effects

// Fiery Conclusion — Instant {1}{R} (Ravnica: City of Guilds, #5600):
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Fiery Conclusion deals 5 damage to target creature."
//
// Bone Splinters' sacrifice-a-creature additional cost (paid as the spell is
// cast, CR 601.2h) with Lightning Axe's 5 damage to the one target. A
// target gone by resolution is a countered spell (CR 608.2b); the sacrifice
// stays paid.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "91af3c81-615a-4b89-acf2-34c72001ebbd",
		Name:           "Fiery Conclusion",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		Targets:        TargetCreature("target creature"),
		Purpose:        ForTargets(DamageToTarget(0, 5)),
		OnResolve:      damageToFirstTarget(5),
	})
}
