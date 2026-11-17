package effects

// Killer Bees — Creature — Insect {1}{G}{G}, 0/1 (Legends, #5639):
//
//	"Flying
//	 {G}: This creature gets +1/+1 until end of turn."
//
// The shared legSelfPump row: no tap in the cost, so it can be activated
// any number of times, summoning sick or not, and the boost ends in the
// cleanup step (CR 514.2). Flying is an engine keyword (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5da1f1af-d2e5-4e14-914b-d93f15626636",
		Name:            "Killer Bees",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []ActivatedAbility{legSelfPump("{G}", 1, "This creature", "Killer Bees")},
	})
}
