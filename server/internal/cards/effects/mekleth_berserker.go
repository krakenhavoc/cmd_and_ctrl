package effects

// Mek'leth Berserker — Creature — Klingon Berserker {3}{R}, 4/3 (Star Trek Commander, #4974):
//
//	"Menace (This creature can't be blocked except by two or more creatures.)"
//
// The keyword is the whole card: PrintedKeywords feeds the synthesized
// layer-6 static, and the engine's combat rules read it from there.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7063eb2b-81a4-4599-b1db-edf861e39bfb",
		Name:            "Mek'leth Berserker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
	})
}
