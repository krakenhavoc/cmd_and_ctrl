package effects

// Obelisk of Bant — Artifact {3} (Shards of Alara, #5588):
//
//	"{T}: Add {G}, {W}, or {U}."
//
// A three-colour mana rock on the tri-lands' b03TriMana row: one tap
// mana ability (CR 605.1a) that adds one mana of a colour its controller
// picks.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "3f96dfd3-975a-4595-95c1-d2f04a86cb37",
		Name:          "Obelisk of Bant",
		Completeness:  CompletenessFull,
		ManaAbilities: []ManaAbility{b03TriMana("G", "W", "U")},
	})
}
