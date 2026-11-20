package effects

// Volcanic Upheaval — Instant {3}{R} (Battle for Zendikar, #5547):
//
//	"Destroy target land."
//
// Stone Rain's sentence at instant speed: any land may be the target,
// checked again at resolution (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4d628f52-183a-4a89-b05f-50463a82513d",
		Name:         "Volcanic Upheaval",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target land", Land()),
		OnResolve:    destroyTheTargetPermanent,
	})
}
