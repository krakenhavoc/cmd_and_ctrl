package effects

// Aerathi Berserker — Creature — Human Berserker {2}{R}{R}{R}, 2/4 (Legends, #5769):
//
//	"Rampage 3 (Whenever this creature becomes blocked, it gets +3/+3
//	 until end of turn for each creature blocking it beyond the first.)"
//
// Rampage is the engine's keyword trigger (CR 702.23, game/rampage.go),
// which counts the creatures blocking it as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5cb495a2-c683-4066-b6ee-d0b7d8843cb9",
		Name:            "Aerathi Berserker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"rampage 3"},
	})
}
