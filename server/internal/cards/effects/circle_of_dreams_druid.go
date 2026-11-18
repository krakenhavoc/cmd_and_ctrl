package effects

// Circle of Dreams Druid — Creature — Elf Druid {G}{G}{G}, 2/1:
//
//	"{T}: Add {G} for each creature you control."
//
// Gaea's Cradle's mana ability on a creature: the count is read as the
// ability resolves (CR 605.3b), the Druid itself included, and a creature
// with summoning sickness can't pay the {T} (CR 302.6), which the engine
// supplies.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1c857fa4-e823-494d-9ea4-7eb9e719c9b6",
		Name:         "Circle of Dreams Druid",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:           ManaAbilityCost{Tap: true},
			ProducedAmount: ProducedPerPermanent("G", MatchCreature),
			Label:          "Add {G} for each creature you control",
		}},
	})
}
