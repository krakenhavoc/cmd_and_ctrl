package effects

// Rofellos, Llanowar Emissary — Legendary Creature — Elf Druid {G}{G},
// 2/1 (Urza's Destiny, #5621):
//
//	"{T}: Add {G} for each Forest you control."
//
// Magus of the Coffers' count with no mana in the cost: a mana ability
// (CR 605.1a) that adds one {G} per permanent with the Forest subtype
// its controller controls as it resolves, basic or not. On a creature
// the {T} needs it to have been under its controller's control since
// their most recent turn began (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3015882b-4897-4d2c-8e33-6731c85a0d03",
		Name:         "Rofellos, Llanowar Emissary",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:           ManaAbilityCost{Tap: true},
			ProducedAmount: ProducedPerPermanent("G", MatchLandSubtype("Forest")),
			Label:          "Add {G} for each Forest you control",
		}},
	})
}
