package effects

// Akroan Sergeant — Creature — Human Soldier {2}{R}, 2/2 (Magic Origins, #5549):
//
//	"First strike (This creature deals combat damage before creatures without first strike.)
//	 Renown 1 (When this creature deals combat damage to a player, if it isn't renowned, put a +1/+1 counter on it and it becomes renowned.)"
//
// Keywords only: first strike, and renown, which is the engine's keyword
// trigger (game/renown.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8a206c6f-4f6f-4c31-9dab-d73a57db989c",
		Name:            "Akroan Sergeant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "renown 1"},
	})
}
