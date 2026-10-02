package effects

// Circle of Despair — Enchantment {1}{W}{B}:
//
//	"{1}, Sacrifice a creature: The next time a source of your choice would deal damage to any target this turn, prevent that damage."
//
// ADR 0107 §6 (#1860): the one-use shield against the next instance of
// damage from a source chosen as it resolves (CR 615.8,
// 609.7a).
// "Any target" is a target (CR 115.4); the creature is sacrificed as
// part of the cost.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "4ad57fbf-69f3-4367-b061-ed8e0b83b9af",
		Name:         "Circle of Despair",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{nextDamageShieldRow(
			"{1}, Sacrifice a creature: The next time a source of your choice would deal damage to any target this turn, prevent that damage.",
			Plus(ManaCost("{1}"), SacrificeACreature()), TargetAny(), PreventNextDamageFromChosenSource(ShieldTheTarget))},
	})
}
