package effects

// Captain's Defense — Instant {W} (Marvel Super Heroes Commander, #4455):
//
//	"Target blocking creature gets +2/+2 until end of turn.
//	 Draw a card."
//
// The target is a blocking creature (CR 506.4). If it is an illegal target
// on resolution the spell does nothing, so no card is drawn (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c3f83115-398f-4738-9bf2-b37bd363e9f6",
		Name:         "Captain's Defense",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target blocking creature", BlockingCreature()),
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2)),
		OnResolve:    pumpBlockerTwoThenDraw("Captain's Defense — +2/+2"),
	})
}
