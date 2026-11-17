package effects

// Dauthi Embrace — Enchantment {2}{B} (Tempest, #5628):
//
//	"{B}{B}: Target creature gains shadow until end of turn. (It can
//	 block or be blocked by only creatures with shadow.)"
//
// Keen Glidemaster's targeted grant with shadow, through the shared
// ebTargetCreatureGainsUntilEOT body: a layer-6 grant on the target
// (CR 613.1f) that ends in the cleanup step (CR 514.2). Any creature may
// be the target, and the ability can be activated any number of times
// at instant speed. A target gone in response does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b19a7ae7-3af5-4aa6-80e1-8888054427de",
		Name:         "Dauthi Embrace",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{B}{B}: Target creature gains shadow until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "shadow")),
			Cost:    ManaCost("{B}{B}"),
			Targets: TargetCreature("target creature"),
			Effect:  ebTargetCreatureGainsUntilEOT("shadow", "Dauthi Embrace — shadow until end of turn"),
		}},
	})
}
