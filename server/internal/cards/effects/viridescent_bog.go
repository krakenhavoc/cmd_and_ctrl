package effects

// Viridescent Bog — Land:
//
//	"{1}, {T}: Add {B}{G}."
//
// A filter-style land in the Signet cycle's shape — a Mana cost
// component ({1}) alongside the tap, producing two fixed colours
// rather than a pipe pick. See signets.go for the mechanical notes
// (the {1} is a real cost, validated before the tap; no auto-tap
// plans it).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6bd6d259-1af7-4dff-a79c-48a616d2a36e",
		Name:         "Viridescent Bog",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, Mana: "{1}"},
			Produced: "{B}{G}",
			Label:    "{1}, {T}: Add {B}{G}",
		}},
	})
}
