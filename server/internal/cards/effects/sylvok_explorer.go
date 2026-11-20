package effects

// Sylvok Explorer — Creature — Human Druid {1}{G}, 1/1 (Fifth Dawn, #5604):
//
//	"{T}: Add one mana of any color that a land an opponent controls
//	 could produce."
//
// Fellwar Stone's derived mana ability on a creature: the colours are read
// at activation from what an opponent's land could produce (CR 106.7), {C}
// filtered out, and nothing to derive from produces nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cf7371c8-2fd8-4e70-abde-27495af85011",
		Name:         "Sylvok Explorer",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:                    ManaAbilityCost{Tap: true},
			DerivedMatch:            DerivedFromOpponentLands(),
			DerivedColorsOnly:       true,
			DerivesFromOtherSources: true,
			Label:                   "Add one mana of any color an opponent's land could produce",
		}},
	})
}
