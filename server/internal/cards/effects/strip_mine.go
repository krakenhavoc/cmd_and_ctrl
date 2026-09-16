package effects

// Strip Mine — Land (EDHREC rank 524):
//
//	"{T}: Add {C}.
//	 {T}, Sacrifice this land: Destroy target land."
//
// The land-destruction land. A tap-and-sacrifice activated ability
// with a land target clause; the cost is paid at announce, so the
// Mine is gone before the ability resolves and a response that
// saves the target leaves the Mine spent — as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d21a89eb-7c5b-459a-acc7-12b20b13bf79",
		Name:         "Strip Mine",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice this land: Destroy target land.",
			Cost:    Plus(TapCost(), SacrificeThis()),
			Targets: TargetPermanent("target land", Land()),
			Effect:  destroyChosenPermanent,
		}},
	})
}
