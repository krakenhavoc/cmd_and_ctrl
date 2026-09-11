package effects

// Ornithopter of Paradise — Artifact Creature — Thopter {2}, 0/2
// (EDHREC rank 221):
//
//	"Flying
//	 {T}: Add one mana of any color."
//
// Birds of Paradise as a colourless artifact: the same five-colour
// pipe (narrowed to the commander's identity at activation, like
// the Birds and every Treasure), a 0/2 body, and a mana cost any
// deck can pay. Flying is printed, so it rides PrintedKeywords and
// the combat engine honours it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3a940dfa-a026-4969-8981-ef90cdfbe9ef",
		Name:            "Ornithopter of Paradise",
		PrintedKeywords: []string{"flying"},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
