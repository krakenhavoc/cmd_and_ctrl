package effects

// Rhox Maulers — Creature — Rhino Soldier {4}{G}, 4/4 (Magic Origins, #5549):
//
//	"Trample (This creature can deal excess combat damage to the player or planeswalker it's attacking.)
//	 Renown 2 (When this creature deals combat damage to a player, if it isn't renowned, put two +1/+1 counters on it and it becomes renowned.)"
//
// Keywords only: trample, and renown 2, the engine's keyword trigger
// (game/renown.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "019461d9-63ae-43d3-a9d4-bd745d0cf46c",
		Name:            "Rhox Maulers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "renown 2"},
	})
}
