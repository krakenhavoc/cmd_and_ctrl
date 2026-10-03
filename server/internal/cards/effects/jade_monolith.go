package effects

// Jade Monolith — Artifact {4}:
//
//	"{1}: The next time a source of your choice would deal damage to
//	 target creature this turn, that source deals that damage to you
//	 instead."
//
// ADR 0108 §9 (#1905): a "next time" redirection from a source chosen as
// the ability resolves (CR 609.7a), from the target to the ability's
// controller.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "1e105ab7-fb10-4cfd-ac2f-5e11488cf1b0",
		Name:         "Jade Monolith",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			redirectRow("{1}: The next time a source of your choice would deal damage to target creature this turn, that source deals that damage to you instead.",
				ManaCost("{1}"), TargetCreature("target creature"),
				RedirectDamage{Choose: true, Protect: ShieldTheTarget, Next: true, To: RedirectToYou}),
		},
	})
}
