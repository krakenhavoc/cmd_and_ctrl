package effects

// Iridescent Angel — Creature — Angel {5}{W}{U}, 4/4 (Odyssey, #5613):
//
//	"Flying, protection from each color"
//
// "Protection from each color" is shorthand for five protection
// abilities, one per colour (CR 702.16h), so the printed keywords are
// flying (CR 702.9b) and protection from white, blue, black, red and
// green. Only a colorless spell, ability or blocker gets past them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "28fa9211-96bb-4fef-81bf-8f58114a8df7",
		Name:            "Iridescent Angel",
		Completeness:    CompletenessFull,
		PrintedKeywords: append([]string{"flying"}, odyProtectionFromEachColor()...),
	})
}
