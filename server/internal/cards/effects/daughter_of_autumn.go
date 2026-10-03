package effects

// Daughter of Autumn — Legendary Creature — Avatar {2}{G}{G}, 2/4:
//
//	"{W}: The next 1 damage that would be dealt to target white creature
//	 this turn is dealt to Daughter of Autumn instead."
//
// ADR 0108 §9 (#1905): a 1-point charged redirection (CR 615.7) from the
// target to Daughter of Autumn as it is now.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "45f3b9f2-3fce-4f91-bcbd-de069e5f9e4c",
		Name:         "Daughter of Autumn",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			redirectRow("{W}: The next 1 damage that would be dealt to target white creature this turn is dealt to Daughter of Autumn instead.",
				ManaCost("{W}"), TargetCreature("target white creature", OfColor("W")),
				RedirectDamage{Protect: ShieldTheTarget, Amount: 1, To: RedirectToThis}),
		},
	})
}
