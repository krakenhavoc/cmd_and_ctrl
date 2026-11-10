package effects

// Sear — Instant {1}{R} (Lorwyn Eclipsed, #4458):
//
//	"Sear deals 4 damage to target creature or planeswalker."
//
// Molten Exhale's body without the behold: 4 damage to one target
// creature or planeswalker (CR 120.3c, 120.3e). If the target is illegal
// as Sear resolves, Sear does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "23b11fdc-61b5-4f5e-ad80-f81b729fc91d",
		Name:         "Sear",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		Purpose:      ForTargets(DamageToTarget(0, 4)),
		OnResolve:    damageToFirstTarget(4),
	})
}
