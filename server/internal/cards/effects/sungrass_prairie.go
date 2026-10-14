package effects

// Sungrass Prairie — Land:
//
//	"{1}, {T}: Add {G}{W}."
//
// The tap-and-a-tax dual, Selesnya half. See darkwater_catacombs.go for
// the shape's mechanics; this is Skycloud Expanse's body with green and
// white.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0a28dff0-2bd6-4105-b73a-b6c4735833fd",
		Name:         "Sungrass Prairie",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced: "{G}{W}",
				Label:    "{1}, {T}: Add {G}{W}",
			},
		},
	})
}
