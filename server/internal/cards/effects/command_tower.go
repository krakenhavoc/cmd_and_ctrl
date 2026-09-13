package effects

// Command Tower — Land:
//
//	"{T}: Add one mana of any color in your commander's color
//	identity."
//
// The single most-played card in Commander, and free to express:
// the identity narrowing already exists for Arcane Signet, so this
// is the same pipe set with the same engine-side filter, on a land.
//
// A land with a catalog entry declaring ManaAbilities bypasses the
// synthetic basic-land ability the engine derives from TypeLine —
// which is what we want here, since "Land" alone would otherwise
// produce nothing.
func init() {
	Register(Spec{
		OracleID:     "0895c9b7-ae7d-4bb3-af17-3b75deb50a25",
		Name:         "Command Tower",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color in your commander's color identity",
		}},
	})
}
