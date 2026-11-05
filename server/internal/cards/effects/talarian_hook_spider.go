package effects

// Talarian Hook Spider — Creature — Spider {1}{G}, 2/2 (Star Trek Commander, #4974):
//
//	"Reach (This creature can block creatures with flying.)"
//
// The keyword is the whole card: PrintedKeywords feeds the synthesized
// layer-6 static, and the engine's combat rules read it from there.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "369f086f-3497-496a-94a8-0faccd1d5931",
		Name:            "Talarian Hook Spider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
