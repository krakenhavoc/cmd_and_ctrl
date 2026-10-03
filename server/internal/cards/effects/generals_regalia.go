package effects

// General's Regalia — Artifact {3}:
//
//	"{3}: The next time a source of your choice would deal damage to you
//	 this turn, that damage is dealt to target creature you control
//	 instead."
//
// ADR 0108 §9 (#1905): a "next time" redirection from a source chosen as
// the ability resolves (CR 609.7a) to the target, chosen as it was
// activated.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "cceba6f3-b1c0-45ee-826c-b871d24c7208",
		Name:         "General's Regalia",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			redirectRow("{3}: The next time a source of your choice would deal damage to you this turn, that damage is dealt to target creature you control instead.",
				ManaCost("{3}"), TargetCreature("target creature you control", YouControl()),
				RedirectDamage{Choose: true, Protect: ShieldYou, Next: true, To: RedirectToClause(0)}),
		},
	})
}
