package effects

// Indatha Crystal — Artifact {3} (Ikoria: Lair of Behemoths, #5512):
//
//	"{T}: Add {W}, {B}, or {G}.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// A mana rock for its wedge, declared like the Triomes' mana ability,
// and cycling from hand (CR 702.29a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3617c4ec-dc70-4fa8-94f3-fbc48abcaeea",
		Name:         "Indatha Crystal",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|B|G}",
			Label:    "{T}: Add {W}, {B}, or {G}",
		}},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
