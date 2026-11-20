package effects

// Markov Patrician — Creature — Vampire {2}{B}, 3/1 (Innistrad, #5572):
//
//	"Lifelink (Damage dealt by this creature also causes you to gain
//	 that much life.)"
//
// Lifelink is an engine keyword carried as printed data: damage it
// deals also gains its controller that much life (CR 702.15b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d56602f3-64df-4c68-a4a3-95c37c021e72",
		Name:            "Markov Patrician",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
}
