package effects

// Lightning Stormkin — Creature — Elemental Wizard {U}{R}, 2/2 (Core Set 2020, #5518):
//
//	"Flying
//	 Haste (This creature can attack and {T} as soon as it comes under your
//	 control.)"
//
// Two printed keywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "126844f5-3a95-48cb-80f2-d7e4b0005350",
		Name:            "Lightning Stormkin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
	})
}
