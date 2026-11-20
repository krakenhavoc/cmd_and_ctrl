package effects

// Banners Raised — {R} Instant (Avacyn Restored, #5570):
//
//	"Creatures you control get +1/+0 until end of turn."
//
// Scare Tactics' text: the creatures are those its controller
// controls as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b60d83ea-18bf-4c64-b4ee-981d54629362",
		Name:         "Banners Raised",
		Completeness: CompletenessFull,
		OnResolve:    creaturesYouControlGetUntilEOT(1, 0, "Banners Raised — creatures you control get +1/+0"),
	})
}
