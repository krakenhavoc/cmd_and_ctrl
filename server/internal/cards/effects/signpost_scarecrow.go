package effects

// Signpost Scarecrow — Artifact Creature — Scarecrow {4}, 2/4 (Throne of Eldraine, #5516):
//
//	"Vigilance
//	 {2}: Add one mana of any color."
//
// Crossroads Candleguide's filter: a mana ability with a mana cost and no {T}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5b6c6dab-0687-4297-b215-89e699d65681",
		Name:            "Signpost Scarecrow",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		ManaAbilities: []ManaAbility{
			samiAnyColorMana(ManaAbilityCost{Mana: "{2}"}, "{2}: Add one mana of any color"),
		},
	})
}
