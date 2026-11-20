package effects

// Fearless Halberdier — Creature — Human Warrior {2}{R}, 3/2 (Guilds of Ravnica, #5525):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2ef2f417-2061-4796-974a-92f800ef05e4",
		Name:         "Fearless Halberdier",
		Completeness: CompletenessFull,
	})
}
