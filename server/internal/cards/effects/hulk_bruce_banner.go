package effects

// Hulk, Bruce Banner — Legendary Creature — Gamma Berserker Hero
// {3}{R}{R}, 7/4 (Marvel Super Heroes Commander, #4455):
//
//	"Trample"
//
// A keyword creature with nothing else on it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1f14d336-71ef-4e3f-b8e0-42585af8db40",
		Name:            "Hulk, Bruce Banner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
