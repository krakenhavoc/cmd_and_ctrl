package effects

// Humongulus — Creature — Homunculus {4}{U}, 2/5 (Ravnica Allegiance,
// #5521):
//
//	"Hexproof (This creature can't be the target of spells or abilities
//	 your opponents control.)"
//
// Hexproof is an engine keyword carried as printed data (CR 702.11b): an
// opponent's spell or ability can't target it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "395922cc-d524-4241-a7d8-9c7959f0b9db",
		Name:            "Humongulus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
	})
}
