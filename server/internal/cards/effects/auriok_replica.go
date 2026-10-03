package effects

// Auriok Replica — Artifact Creature — Cleric {3}:
//
//	"{W}, Sacrifice this creature: Prevent all damage a source of your choice would deal to you this turn."
//
// ADR 0108 §7 (#1904): a shield against a source chosen as the ability
// resolves (CR 609.7a), protecting its controller from every instance of
// that source's damage this turn, not only the next.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "783a62c0-3b06-4250-a1cb-5cf4042611aa",
		Name:         "Auriok Replica",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{sourceShieldRow(
			"{W}, Sacrifice this creature: Prevent all damage a source of your choice would deal to you this turn.",
			Plus(ManaCost("{W}"), SacrificeThis()), nil,
			PreventDamageFromChosenSource(ShieldYou))},
	})
}
