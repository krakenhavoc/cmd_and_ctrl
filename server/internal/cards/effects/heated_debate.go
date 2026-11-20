package effects

// Heated Debate — Instant {2}{R} (Strixhaven: School of Mages, #5502):
//
//	"This spell can't be countered. (This includes by the ward ability.)
//	 Heated Debate deals 4 damage to target creature or planeswalker."
//
// Sear's 4 damage to a creature or planeswalker, with Abrupt Decay's
// can't-be-countered flag, which a ward trigger's counter respects too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "06c3e85c-17d8-44e1-a0f8-07ef3f1624cb",
		Name:            "Heated Debate",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		Targets:         TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		Purpose:         ForTargets(DamageToTarget(0, 4)),
		OnResolve:       damageToFirstTarget(4),
	})
}
