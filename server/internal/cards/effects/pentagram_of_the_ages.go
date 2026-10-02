package effects

// Pentagram of the Ages — Artifact {4}:
//
//	"{4}, {T}: The next time a source of your choice would deal damage to you this turn, prevent that damage."
//
// ADR 0107 §6 (#1860): the one-use shield against the next instance of
// damage from a source chosen as it resolves (CR 615.8,
// 609.7a).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "265204a9-0ce2-4a0a-a9ea-8537cd94b079",
		Name:         "Pentagram of the Ages",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{nextDamageShieldRow(
			"{4}, {T}: The next time a source of your choice would deal damage to you this turn, prevent that damage.",
			Plus(ManaCost("{4}"), TapCost()), nil, PreventNextDamageFromChosenSource(ShieldYou))},
	})
}
