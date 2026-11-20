package effects

// Cold-Water Snapper — Creature — Turtle {5}{U}, 4/5 (Dominaria, #5530):
//
//	"Hexproof (This creature can't be the target of spells or abilities your opponents control.)"
//
// One printed keyword: an opponent's spell or ability can't target it (CR 702.11b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "457e4a49-c8fe-4e1d-a4b4-7a880a47f58d",
		Name:            "Cold-Water Snapper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
	})
}
