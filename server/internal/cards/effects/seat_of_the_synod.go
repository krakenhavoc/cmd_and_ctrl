package effects

// Seat of the Synod — Artifact Land:
//
//	"{T}: Add {U}."
//
// The whole card. Its value is the type line, not the text: it is a
// land that is also an ARTIFACT, which is what turns on metalcraft,
// affinity, Foundry Inspector, Inventors' Fair and every other
// artifact count on the table — all of which read the printed type
// line through Card.IsArtifact() and need nothing declared here.
//
// Registered because an unregistered nonbasic land taps for nothing:
// the engine's synthetic mana ability (ManaAbilitiesForCard →
// basicLandColor) only fires for lands carrying the BASIC supertype,
// and "Artifact Land" has no basic land type to derive from.
//
// No enters-tapped clause — the Mirrodin artifact lands famously
// have none, which is the other half of why they are played.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID: "39451b4d-cd7a-40da-b457-cb51b609173f",
		Name:     "Seat of the Synod",
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
	})
}
