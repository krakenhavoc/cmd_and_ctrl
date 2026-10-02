package effects

// Sanctum Guardian — Creature — Human Cleric {1}{W}{W}:
//
//	"Sacrifice this creature: The next time a source of your choice would deal damage to any target this turn, prevent that damage."
//
// ADR 0107 §6 (#1860): the one-use shield against the next instance of
// damage from a source chosen as it resolves (CR 615.8,
// 609.7a).
// "Any target" is a target (CR 115.4); the Guardian is sacrificed as part
// of the cost, so it may protect anything but itself.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "eb7fe094-2237-40c8-b831-93498c93b768",
		Name:         "Sanctum Guardian",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{nextDamageShieldRow(
			"Sacrifice this creature: The next time a source of your choice would deal damage to any target this turn, prevent that damage.",
			SacrificeThis(), TargetAny(), PreventNextDamageFromChosenSource(ShieldTheTarget))},
	})
}
