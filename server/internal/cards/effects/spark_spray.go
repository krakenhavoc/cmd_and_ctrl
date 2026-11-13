package effects

// Spark Spray — Instant {R} (Scourge):
//
//	"Spark Spray deals 1 damage to any target.
//	 Cycling {R} ({R}, Discard this card: Draw a card.)"
//
// One damage to any target, and cycling as the keyword's own hand ability
// (CR 702.29a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c37f1921-7ac3-4185-8579-1dfdfe647ea2",
		Name:         "Spark Spray",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(DamageToTarget(0, 1)),
		Targets:      TargetAny(),
		OnResolve:    damageToFirstTarget(1),
		Activated:    []ActivatedAbility{Cycling("{R}")},
	})
}
