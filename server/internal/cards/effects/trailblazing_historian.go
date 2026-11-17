package effects

// Trailblazing Historian — Creature — Human Shaman 1/3 {1}{R} (March of the Machine, S82):
//
//	"Haste
//	 {T}: Another target creature gains haste until end of turn."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "399770b9-eab7-4512-aa1d-65c5bd81094f",
		Name:            "Trailblazing Historian",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Another target creature gains haste until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "haste")),
			Cost:    TapCost(),
			Targets: Another(TargetCreature("another target creature")),
			Effect:  ebTargetCreatureGainsUntilEOT("haste", "Trailblazing Historian — haste until end of turn"),
		}},
	})
}
