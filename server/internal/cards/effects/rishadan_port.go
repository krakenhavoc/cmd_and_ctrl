package effects

// Rishadan Port — Land:
//
//	"{T}: Add {C}.
//	 {1}, {T}: Tap target land."
//
// The second ability is the shared tapChosenPermanent body; the target must
// still be a land as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f3e8dc56-2810-474e-a6a7-9c3555f94ae9",
		Name:         "Rishadan Port",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Tap target land.",
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetPermanent("target land", Land()),
			Effect:  tapChosenPermanent,
		}},
	})
}
