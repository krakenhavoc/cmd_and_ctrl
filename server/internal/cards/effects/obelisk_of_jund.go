package effects

// Obelisk of Jund — Artifact {3} (Shards of Alara, #5588):
//
//	"{T}: Add {B}, {R}, or {G}."
//
// A three-colour mana rock on the tri-lands' b03TriMana row: one tap
// mana ability (CR 605.1a) that adds one mana of a colour its controller
// picks.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "eb0be178-55f9-4743-8128-34bc86f134ec",
		Name:          "Obelisk of Jund",
		Completeness:  CompletenessFull,
		ManaAbilities: []ManaAbility{b03TriMana("B", "R", "G")},
	})
}
