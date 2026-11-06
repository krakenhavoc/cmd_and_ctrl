package effects

// Capital Guard — Creature — Human Soldier {1}{R}, 2/2 (Avatar: The Last
// Airbender Eternal, #5075):
//
// No rules text: a vanilla creature, registered so it does not wear the
// "unimplemented" badge.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aff20717-265d-4214-a8f5-f520bb243cb5",
		Name:         "Capital Guard",
		Completeness: CompletenessFull,
	})
}
