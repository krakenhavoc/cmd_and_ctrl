package effects

// Fear of Exposure — Enchantment Creature — Nightmare {2}{G} (#3868):
//
//	"As an additional cost to cast this spell, tap two untapped
//	 creatures and/or lands you control.
//	 Trample"
//
// The additional cost is New Blood's (TapUntappedCost) over creatures and
// lands; a land tapped to pay it is not tapped for the {2}{G}, and a
// creature that entered this turn may pay (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "abec56eb-7b3d-4bd8-93c7-bd7049c859c3",
		Name:            "Fear of Exposure",
		Completeness:    CompletenessFull,
		AdditionalCost:  TapUntappedCost(2, "two untapped creatures and/or lands you control", Or(Creature(), Land())),
		PrintedKeywords: []string{"trample"},
	})
}
