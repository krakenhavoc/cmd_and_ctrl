package effects

// Streaking Oilgorger — Creature — Vampire {4}{B}, 3/3:
//
//	"Flying, haste
//	 Start your engines!
//	 Max speed — This creature has lifelink."
//
// ADR 0138 (#2711).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c163a95e-9414-4b8c-b5a2-1e7674530404",
		Name:            "Streaking Oilgorger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste", StartYourEngines},
		Static:          MaxSpeedSelfKeywords("lifelink"),
	})
}
