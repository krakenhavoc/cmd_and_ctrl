package effects

// Vector Asp — Artifact Creature — Phyrexian Snake {1}, 1/1 (Scars of
// Mirrodin, #5577):
//
//	"{B}: This creature gains infect until end of turn."
//
// A self grant until end of turn (somB1SelfGainsUntilEOT); infect is the
// engine's damage tail (CR 702.90), read as damage is dealt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1950067a-cc57-42a8-b1ee-f03e1b50591b",
		Name:         "Vector Asp",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{somB1SelfGainsUntilEOT("Vector Asp", "{B}", "infect")},
	})
}
