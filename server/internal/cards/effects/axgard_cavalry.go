package effects

// Axgard Cavalry — Creature — Dwarf Berserker {1}{R}, 2/2 (Foundations,
// #4465):
//
//	"{T}: Target creature gains haste until end of turn."
//
// Taxi Driver's haste grant with only {T} as its cost. The Cavalry's own
// {T} needs it to have been under your control since your most recent
// turn began (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6a92b011-fd01-4b6a-a17c-73ac1967ff98",
		Name:         "Axgard Cavalry",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target creature gains haste until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "haste")),
			Cost:    TapCost(),
			Targets: TargetCreature("target creature"),
			Effect:  ebTargetCreatureGainsUntilEOT("haste", "Axgard Cavalry — haste until end of turn"),
		}},
	})
}
