package effects

// Essence Burn — Instant {1}{R}:
//
//	"Essence Burn deals 5 damage to target black or green creature or
//	 planeswalker. If that permanent would die this turn, exile it
//	 instead."
//
// Lava Coil's sentence on a colour-restricted creature-or-planeswalker
// target. The replacement is the spell's, so the permanent is marked
// whether or not the damage is dealt (ADR 0108 §1).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "f1216f36-00fd-4d03-85c1-f4910a02d4d0",
		Name:         "Essence Burn",
		Completeness: CompletenessFull,
		Targets: TargetPermanent("target black or green creature or planeswalker",
			And(Or(Creature(), Planeswalker()), Or(OfColor("B"), OfColor("G")))),
		Purpose:   ForTargets(DamageToTarget(0, 5)),
		OnResolve: damageFirstTargetExileIfItDies(fixedAmount(5)),
	})
}
