package effects

// Ancient Den — Artifact Land (EDHREC rank 491):
//
//	"{T}: Add {W}."
//
// A Plains that is also an artifact — the affinity and Cranial
// Plating land. It needs a Spec at all because the engine's
// synthetic land ability requires the BASIC supertype; an
// unregistered Ancient Den would tap for nothing. The artifact half
// is the printed type line, which every artifact-counting card
// (Storm-Kiln Artist) reads.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "02f16726-f2f6-4943-b71a-93f8e26251d3",
		Name:     "Ancient Den",
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
	})
}
