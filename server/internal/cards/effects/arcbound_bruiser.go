package effects

// Arcbound Bruiser — Artifact Creature — Golem {5}, 0/0 (Darksteel,
// #5605):
//
//	"Modular 3 (This creature enters with three +1/+1 counters on it.
//	 When it dies, you may put its +1/+1 counters on target artifact
//	 creature.)"
//
// Modular is the engine's keyword (#2012, game/modular.go, CR 702.43a):
// the printed token carries both halves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "51577a8b-28d6-4f9e-8d3f-6a1380270fab",
		Name:            "Arcbound Bruiser",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"modular 3"},
	})
}
