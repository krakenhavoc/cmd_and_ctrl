package effects

// Soul of the Rapids — Creature — Elemental {3}{U}{U}, 3/2 (Rivals of Ixalan, #5531):
//
//	"Flying
//	 Hexproof (This creature can't be the target of spells or abilities
//	 your opponents control.)"
//
// Flying and hexproof are engine keywords carried as printed data: only a
// flier or a creature with reach blocks it, and an opponent's spell or
// ability can't target it (CR 702.11b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "011b2243-b8be-471f-9d8f-824a40e599df",
		Name:            "Soul of the Rapids",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "hexproof"},
	})
}
