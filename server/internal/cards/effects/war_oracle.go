package effects

// War Oracle — Creature — Human Cleric {2}{W}{W}, 3/3 (Magic Origins, #5549):
//
//	"Lifelink (Damage dealt by this creature also causes you to gain that much life.)
//	 Renown 1 (When this creature deals combat damage to a player, if it isn't renowned, put a +1/+1 counter on it and it becomes renowned.)"
//
// Keywords only: lifelink, and renown, the engine's keyword trigger
// (game/renown.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "aa26fb50-1b53-488e-90d1-b0c7af273622",
		Name:            "War Oracle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink", "renown 1"},
	})
}
