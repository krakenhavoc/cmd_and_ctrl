package effects

// Arcbound Lancer — Artifact Creature — Beast {7}, 0/0 (Darksteel, #5605):
//
//	"First strike
//	 Modular 4 (This creature enters with four +1/+1 counters on it. When
//	 it dies, you may put its +1/+1 counters on target artifact
//	 creature.)"
//
// First strike and modular are both engine keywords (modular: #2012,
// game/modular.go, CR 702.43a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fd6ffa00-170b-46d2-810a-f255de62b9d7",
		Name:            "Arcbound Lancer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "modular 4"},
	})
}
