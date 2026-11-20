package effects

// Cloaked Siren — Creature — Siren {3}{U}, 3/2 (Journey into Nyx, #5558):
//
//	"Flash (You may cast this spell any time you could cast an
//	 instant.)
//	 Flying"
//
// Wind Strider's printed keywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d38a33ae-59d1-4d08-8e14-5fa5a8c3762d",
		Name:            "Cloaked Siren",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
	})
}
