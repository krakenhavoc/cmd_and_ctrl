package effects

// Arcbound Worker — Artifact Creature — Construct {1}, 0/0 (Darksteel,
// #5605):
//
//	"Modular 1 (This creature enters with a +1/+1 counter on it. When it
//	 dies, you may put its +1/+1 counters on target artifact creature.)"
//
// Modular is the engine's keyword (#2012, game/modular.go, CR 702.43a):
// the printed token carries both halves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6d9af084-cd5d-4242-b0f0-24a36d0a75d4",
		Name:            "Arcbound Worker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"modular 1"},
	})
}
