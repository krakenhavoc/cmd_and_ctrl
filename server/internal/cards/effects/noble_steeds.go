package effects

// Noble Steeds — Enchantment {2}{W} (Alliances, #5634):
//
//	"{1}{W}: Target creature gains first strike until end of turn."
//
// A layer 6 grant (CR 613.1f) to the target that ends at cleanup
// (CR 514.2). A target that has left by the time it resolves gets nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d6005e47-d545-4c16-b54a-bfe1dde61a98",
		Name:         "Noble Steeds",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}: Target creature gains first strike until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "first strike")),
			Cost:    ManaCost("{1}{W}"),
			Targets: TargetCreature("target creature"),
			Effect:  ebTargetCreatureGainsUntilEOT("first strike", "Noble Steeds — first strike until end of turn"),
		}},
	})
}
