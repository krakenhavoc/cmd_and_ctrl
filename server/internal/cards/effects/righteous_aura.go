package effects

// Righteous Aura — Enchantment {1}{W}:
//
//	"{W}, Pay 2 life: The next time a source of your choice would deal damage to you this turn, prevent that damage."
//
// ADR 0107 §6 (#1860): the one-use shield against the next instance of
// damage from a source chosen as it resolves (CR 615.8,
// 609.7a).
// The 2 life is part of the cost (CR 119.4).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "5bf93d4f-be59-4aa7-83d5-7c35d1f07fae",
		Name:         "Righteous Aura",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{nextDamageShieldRow(
			"{W}, Pay 2 life: The next time a source of your choice would deal damage to you this turn, prevent that damage.",
			Plus(ManaCost("{W}"), PayLife(2)), nil, PreventNextDamageFromChosenSource(ShieldYou))},
	})
}
