package effects

// Nimbus Maze — Land:
//
//	"{T}: Add {C}.
//	 {T}: Add {W}. Activate only if you control an Island.
//	 {T}: Add {U}. Activate only if you control a Plains."
//
// The Tainted Isle shape with two gated colours: each coloured ability
// is a `ManaAbility.Condition` reading post-layer land types. The land
// has no land types of its own, so it never satisfies its own gate.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d7e1d4eb-1d4e-460e-9304-7db9ab50ccb5",
		Name:         "Nimbus Maze",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:      ManaAbilityCost{Tap: true},
				Produced:  "{W}",
				Label:     "{T}: Add {W}. Activate only if you control an Island",
				Condition: ControlsAtLeast(1, MatchLandSubtype("Island")),
			},
			{
				Cost:      ManaAbilityCost{Tap: true},
				Produced:  "{U}",
				Label:     "{T}: Add {U}. Activate only if you control a Plains",
				Condition: ControlsAtLeast(1, MatchLandSubtype("Plains")),
			},
		},
	})
}
