package effects

// Hussar Patrol — Creature — Human Knight {2}{W}{U}, 2/4 (Return to Ravnica, #5566):
//
//	"Flash (You may cast this spell any time you could cast an
//	 instant.) / Vigilance"
//
// The printed keywords ride PrintedKeywords; the engine enforces them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bd73581c-c4bd-46cb-8004-75aeafa24288",
		Name:            "Hussar Patrol",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "vigilance"},
	})
}
