package effects

// Darklit Gargoyle — Artifact Creature — Gargoyle {1}{W}, 1/2 (Conflux,
// #5587):
//
//	"Flying
//	 {B}: This creature gets +2/-1 until end of turn."
//
// Flying is a printed keyword; the row is the shared thisGetsRow, ending
// in the cleanup step (CR 514.2). Activating it twice on a 1/2 leaves it
// at 0 toughness, and it dies to the state-based check (CR 704.5f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e48d9efa-c169-4587-86fc-0021f265d69b",
		Name:            "Darklit Gargoyle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []ActivatedAbility{thisGetsRow("{B}", 2, -1, "Darklit Gargoyle")},
	})
}
