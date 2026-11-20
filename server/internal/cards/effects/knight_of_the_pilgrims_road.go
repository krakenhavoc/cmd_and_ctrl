package effects

// Knight of the Pilgrim's Road — Creature — Human Knight {2}{W}, 3/2 (Magic Origins, #5549):
//
//	"Renown 1 (When this creature deals combat damage to a player, if it isn't renowned, put a +1/+1 counter on it and it becomes renowned.)"
//
// Renown only, the engine's keyword trigger (game/renown.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "40999aa6-6424-42e4-9d47-33bee6bbb886",
		Name:            "Knight of the Pilgrim's Road",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"renown 1"},
	})
}
