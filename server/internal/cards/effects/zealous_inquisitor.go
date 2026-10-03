package effects

// Zealous Inquisitor — Creature — Human Cleric {2}{W}, 2/2:
//
//	"{1}{W}: The next 1 damage that would be dealt to this creature this
//	 turn is dealt to target creature instead."
//
// ADR 0108 §9 (#1905): a 1-point charged redirection (CR 615.7) from the
// Inquisitor to the target.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "e85f6251-8801-414b-adc9-5488794e7456",
		Name:         "Zealous Inquisitor",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			redirectRow("{1}{W}: The next 1 damage that would be dealt to this creature this turn is dealt to target creature instead.",
				ManaCost("{1}{W}"), TargetCreature("target creature"),
				RedirectDamage{Protect: ShieldThis, Amount: 1, To: RedirectToClause(0)}),
		},
	})
}
