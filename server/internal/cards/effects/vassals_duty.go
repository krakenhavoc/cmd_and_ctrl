package effects

// Vassal's Duty — Enchantment {3}{W}:
//
//	"{1}: The next 1 damage that would be dealt to target legendary
//	 creature you control this turn is dealt to you instead."
//
// ADR 0108 §9 (#1905): a 1-point charged redirection (CR 615.7) from the
// target to the ability's controller.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "7d84e667-2f14-437c-be4f-161b98d59341",
		Name:         "Vassal's Duty",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			redirectRow("{1}: The next 1 damage that would be dealt to target legendary creature you control this turn is dealt to you instead.",
				ManaCost("{1}"), TargetCreature("target legendary creature you control", Legendary(), YouControl()),
				RedirectDamage{Protect: ShieldTheTarget, Amount: 1, To: RedirectToYou}),
		},
	})
}
