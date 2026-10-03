package effects

// Lancers en-Kor — Creature — Kor Soldier {3}{W}{W}, 3/3:
//
//	"Trample
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
		OracleID:     "ef49dc78-9fd9-4cf8-af10-5a6b8ef7fc57",
		Name:         "Lancers en-Kor",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{enKorRow()},
	})
}
