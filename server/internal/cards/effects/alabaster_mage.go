package effects

// Alabaster Mage — Creature — Human Wizard {1}{W}, 2/1 (Jumpstart):
//
//	"{1}{W}: Target creature you control gains lifelink until end of turn."
//
// No {T} in the cost, so it can be activated with summoning sickness and
// more than once. The target is re-checked on resolution (CR 608.2b): a
// creature that changed control in response gains nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ea7a8acb-f01f-4a69-9b3c-954d8f0d3155",
		Name:         "Alabaster Mage",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}: Target creature you control gains lifelink until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "lifelink")),
			Cost:    ManaCost("{1}{W}"),
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect:  jmpB7FirstLegalTargetGains("Alabaster Mage — lifelink until end of turn", "lifelink"),
		}},
	})
}
