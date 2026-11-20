package effects

// Savai Crystal — Artifact {3} (Ikoria: Lair of Behemoths, #5512):
//
//	"{T}: Add {R}, {W}, or {B}.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// A mana rock for its wedge, declared like the Triomes' mana ability,
// and cycling from hand (CR 702.29a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cd0c1165-673d-4936-8254-deedd0ebbe97",
		Name:         "Savai Crystal",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R|W|B}",
			Label:    "{T}: Add {R}, {W}, or {B}",
		}},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
