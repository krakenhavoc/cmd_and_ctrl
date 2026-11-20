package effects

// Fry — Instant {1}{R} (Core Set 2020, #5518):
//
//	"This spell can't be countered.
//	 Fry deals 5 damage to target creature or planeswalker that's white or
//	 blue."
//
// Spec.CantBeCountered with a colour-gated damage target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3a747893-a514-4689-8486-ecccf654fecb",
		Name:            "Fry",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		Targets: TargetPermanent("target creature or planeswalker that's white or blue",
			And(Or(Creature(), Planeswalker()), Or(OfColor("W"), OfColor("U")))),
		Purpose:   ForTargets(DamageToTarget(0, 5)),
		OnResolve: damageToFirstTarget(5),
	})
}
