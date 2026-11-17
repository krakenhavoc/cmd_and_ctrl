package effects

// Wild Aesthir — Creature — Bird {2}{W}, 1/1 (Alliances, #5634):
//
//	"Flying, first strike
//	 {W}{W}: This creature gets +2/+0 until end of turn. Activate only
//	 once each turn."
//
// Flying and first strike are engine keywords carried as printed data
// (CR 702.9b, CR 702.7b). The boost is a layer 7c change (CR 613.4c) that
// ends at cleanup (CR 514.2); the once-each-turn restriction is counted
// per object (CR 602.5b), so a second activation the same turn is
// refused.
//
// No simplification.
func init() {
	row := thisGetsRow("{W}{W}", 2, 0, "Wild Aesthir")
	row.Label += " Activate only once each turn."
	row.Condition = OncePerTurnActivation(row.Label)
	Register(Spec{
		OracleID:        "dbdf076c-18ce-45e3-8f70-0a32391ea4b6",
		Name:            "Wild Aesthir",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
		Activated:       []ActivatedAbility{row},
	})
}
