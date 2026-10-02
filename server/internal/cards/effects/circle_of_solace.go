package effects

// Circle of Solace — Enchantment {3}{W}:
//
//	"As this enchantment enters, choose a creature type.
//	 {1}{W}: The next time a creature of the chosen type would deal damage to you this turn, prevent that damage."
//
// ADR 0107 §6 (#1860). Nothing is chosen at resolution: the shield is
// against the next instance of damage to you from any creature of the
// type chosen as the enchantment entered (CR 615.8), judged as that
// creature would deal the damage — a changeling is every type
// (CR 702.73a). The type is read as Circle of Solace last existed if it
// has left by resolution (CR 608.2h).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "14ebedd1-7a2d-413b-9b76-8610511292a4",
		Name:         "Circle of Solace",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCreatureTypeAsEnters("Circle of Solace"),
		Activated: []ActivatedAbility{chosenTypeShieldRow(
			"{1}{W}: The next time a creature of the chosen type would deal damage to you this turn, prevent that damage.", ManaCost("{1}{W}"))},
	})
}
