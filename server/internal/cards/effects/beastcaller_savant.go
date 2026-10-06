package effects

// Beastcaller Savant — Creature — Elf Shaman Ally {1}{G}, 1/1:
//
//	"Haste
//	 {T}: Add one mana of any color. Spend this mana only to cast a
//	 creature spell."
//
// The mana carries the CR 106.12 spend restriction, so it pays for a
// creature spell and for nothing else — not an activated ability, not
// a noncreature spell. It is a mana ability (no stack), and as a
// creature's {T} ability it is gated by summoning sickness unless the
// Savant has haste, which it prints.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e28227eb-b3b5-42fa-b597-1fefd2a70186",
		Name:            "Beastcaller Savant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			Produced:     "{W|U|B|R|G}",
			Restrictions: []string{ManaRestrictCast, ManaRestrictType("Creature")},
			Label:        "Add one mana of any color. Spend this mana only to cast a creature spell",
		}},
	})
}
