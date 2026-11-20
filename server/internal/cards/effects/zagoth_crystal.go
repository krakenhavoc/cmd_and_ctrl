package effects

// Zagoth Crystal — Artifact {3} (Ikoria: Lair of Behemoths, #5512):
//
//	"{T}: Add {B}, {G}, or {U}.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// A mana rock for its wedge, declared like the Triomes' mana ability,
// and cycling from hand (CR 702.29a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "87edb566-8929-4961-9c03-d3657a8c6feb",
		Name:         "Zagoth Crystal",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B|G|U}",
			Label:    "{T}: Add {B}, {G}, or {U}",
		}},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
