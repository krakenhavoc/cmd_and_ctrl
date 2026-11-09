package effects

// Sun-Dappled Celebrant — Creature — Treefolk Cleric {4}{W}{W}, 5/6
// (Lorwyn Eclipsed, #4458):
//
//	"Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 Vigilance"
//
// Convoke is Appeal to Eirdu's TapCost; vigilance is printed data.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "27c87371-697e-4ca6-a7f4-d99215d908bc",
		Name:            "Sun-Dappled Celebrant",
		Completeness:    CompletenessFull,
		TapCost:         Convoke(),
		PrintedKeywords: []string{"vigilance"},
	})
}
