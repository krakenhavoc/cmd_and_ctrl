package effects

// Ward of Piety — Enchantment — Aura {1}{W}:
//
//	"Enchant creature
//	 {1}{W}: The next 1 damage that would be dealt to enchanted creature
//	 this turn is dealt to any target instead."
//
// ADR 0108 §9 (#1905): a 1-point charged redirection (CR 615.7) from the
// creature the Aura enchants as the ability resolves to the target. The
// ruling: redirected combat damage is still combat damage, and still the
// original source's.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "0fe4767b-2023-49fc-bc93-b73b43f70b76",
		Name:         "Ward of Piety",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Activated: []ActivatedAbility{
			redirectRow("{1}{W}: The next 1 damage that would be dealt to enchanted creature this turn is dealt to any target instead.",
				ManaCost("{1}{W}"), TargetAny(),
				RedirectDamage{Protect: ShieldEnchantedCreature, Amount: 1, To: RedirectToClause(0)}),
		},
	})
}
