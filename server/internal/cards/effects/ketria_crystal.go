package effects

// Ketria Crystal — Artifact {3} (Ikoria: Lair of Behemoths, #5512):
//
//	"{T}: Add {G}, {U}, or {R}.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// A mana rock for its wedge, declared like the Triomes' mana ability,
// and cycling from hand (CR 702.29a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3ee5d7b6-0263-4239-a5fb-de1e1a220c55",
		Name:         "Ketria Crystal",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G|U|R}",
			Label:    "{T}: Add {G}, {U}, or {R}",
		}},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
