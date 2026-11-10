package effects

// Ancient Ziggurat — Land:
//
//	"{T}: Add one mana of any color. Spend this mana only to cast a
//	 creature spell."
//
// Beastcaller Savant's restricted mana on a land: the mana carries the
// CR 106.12 spend restriction, so it pays for a creature spell and for
// nothing else.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "0baabe39-72ae-47bd-a095-cbf7eb8a6361",
		Name:          "Ancient Ziggurat",
		Completeness:  CompletenessFull,
		ManaAbilities: []ManaAbility{cmdr2B01TapForCreatureOnlyMana()},
	})
}
