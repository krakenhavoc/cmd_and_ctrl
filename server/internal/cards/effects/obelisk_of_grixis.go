package effects

// Obelisk of Grixis — Artifact {3} (Shards of Alara, #5588):
//
//	"{T}: Add {U}, {B}, or {R}."
//
// A three-colour mana rock on the tri-lands' b03TriMana row: one tap
// mana ability (CR 605.1a) that adds one mana of a colour its controller
// picks.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "e193ae18-91a0-4870-b8a1-f8dedf744708",
		Name:          "Obelisk of Grixis",
		Completeness:  CompletenessFull,
		ManaAbilities: []ManaAbility{b03TriMana("U", "B", "R")},
	})
}
