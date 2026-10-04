package effects

// Cascading Cataracts — Land:
//
//	"Indestructible
//	 {T}: Add {C}.
//	 {5}, {T}: Add five mana in any combination of colors."
//
// Five INDEPENDENT any-colour slots (AnyCombinationOfColors), the
// shape Selvala uses, so the controller can answer {W}{U}{B}{R}{G} or
// five of one colour. The ability has no identity narrowing because the
// printed text names none. Indestructible is a printed keyword, as on
// Darksteel Citadel.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d98b4250-3492-4864-9c4c-42db09b3ccd4",
		Name:            "Cascading Cataracts",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true, Mana: "{5}"},
				Produced: AnyCombinationOfColors(5),
				Label:    "{5}, {T}: Add five mana in any combination of colors",
			},
		},
	})
}
