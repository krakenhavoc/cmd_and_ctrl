package effects

// Spirit en-Kor — Creature — Kor Spirit {3}{W}, 2/2:
//
//	"Flying
//	 {0}: The next 1 damage that would be dealt to this creature this
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
		OracleID:     "f066174a-a959-4365-920f-c04506d94a5a",
		Name:         "Spirit en-Kor",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{enKorRow()},
	})
}
