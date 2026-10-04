package effects

// Tainted Peak — Land:
//
//	"{T}: Add {C}.
//	 {T}: Add {B} or {R}. Activate only if you control a Swamp."
//
// The conditional mana ability is Mox Opal's shape: ManaAbility.Condition
// gates activation. "A Swamp" reads the EFFECTIVE land subtypes, so a
// Urborg-ed board enables it and a Blood Moon one does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b2bae7fc-0668-4b34-9cd6-0d80aea52275",
		Name:         "Tainted Peak",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:      ManaAbilityCost{Tap: true},
				Produced:  "{B|R}",
				Label:     "{T}: Add {B} or {R}. Activate only if you control a Swamp",
				Condition: ControlsAtLeast(1, MatchLandSubtype("Swamp")),
			},
		},
	})
}
