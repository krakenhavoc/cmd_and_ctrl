package effects

// Darksteel Ingot — Artifact {3}:
//
//	"Indestructible (Effects that say "destroy" don't destroy this
//	 artifact.)
//	 {T}: Add one mana of any color."
//
// Thornglint Bridge's indestructibility (PrintedKeywords is what the
// destruction paths read) on a Birds of Paradise mana ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a2529491-7389-4cfa-92d2-145eda779603",
		Name:            "Darksteel Ingot",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		ManaAbilities:   []ManaAbility{cmdr2B01TapForAnyColor()},
	})
}
