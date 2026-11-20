package effects

// Akki Drillmaster — Creature — Goblin Shaman {2}{R}, 2/2 (Saviors of
// Kamigawa, #5601):
//
//	"{T}: Target creature gains haste until end of turn."
//
// A layer-6 grant until the cleanup step (CR 514.2); a target that left
// in response is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aa228a01-7de3-4539-9ebb-b3c546f72b41",
		Name:         "Akki Drillmaster",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target creature gains haste until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "haste")),
			Cost:    TapCost(),
			Targets: TargetCreature("target creature"),
			Effect:  neoB07TargetGetsUntilEOT("Akki Drillmaster — haste until end of turn", 0, 0, "haste"),
		}},
	})
}
