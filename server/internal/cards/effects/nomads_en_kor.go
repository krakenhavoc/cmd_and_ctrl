package effects

// Nomads en-Kor — Creature — Kor Nomad Soldier {W}, 1/1:
//
//	"{0}: The next 1 damage that would be dealt to this creature this
//	 turn is dealt to target creature you control instead."
//
// ADR 0108 §9 (#1905): the en-Kor row (enKorRow), a 1-point charged
// redirection (CR 615.7) from this creature to the target. The rulings:
// redirected combat damage is still combat damage; it can redirect to
// itself, which does nothing; and it can be activated as often as you
// like before the damage.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "b8d395a3-0bfe-45c3-bb3a-820d4f235b88",
		Name:         "Nomads en-Kor",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{enKorRow()},
	})
}
