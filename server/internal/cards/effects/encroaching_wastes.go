package effects

// Encroaching Wastes — Land (Magic 2014, #5563):
//
//	"{T}: Add {C}.
//	 {4}, {T}, Sacrifice this land: Destroy target nonbasic land."
//
// Wasteland at {4} more: a colorless land whose sacrifice destroys a nonbasic
// land.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "43144f06-079b-4515-a03a-01ea3e90d586",
		Name:         "Encroaching Wastes",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}, Sacrifice this land: Destroy target nonbasic land.",
			Cost:    Plus(ManaCost("{4}"), TapCost(), SacrificeThis()),
			Targets: TargetPermanent("target nonbasic land", Land(), b10NonbasicLand()),
			Effect:  destroyChosenPermanent,
		}},
	})
}
