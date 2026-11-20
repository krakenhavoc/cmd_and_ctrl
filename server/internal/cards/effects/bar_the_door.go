package effects

// Bar the Door — Instant {2}{W} (Dark Ascension, #5571):
//
//	"Creatures you control get +0/+4 until end of turn."
//
// Steadfastness's team pump with +0/+4: the creatures you control as it
// resolves get it (CR 611.2c); one that arrives later does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1ecd7248-ff28-4e12-b72a-0cb279ec7561",
		Name:         "Bar the Door",
		Completeness: CompletenessFull,
		OnResolve:    creaturesYouControlGetUntilEOT(0, 4, "Bar the Door — creatures you control get +0/+4"),
	})
}
