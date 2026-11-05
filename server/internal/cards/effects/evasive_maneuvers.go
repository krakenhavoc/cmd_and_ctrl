package effects

// Evasive Maneuvers — {W} Instant (Star Trek Commander, #4974):
//
//	"Target blocking creature you control gets +2/+2 until end of turn.
//	 Draw a card."
//
// Captain's Defense with "you control" on the target: the same
// resolution (pumpBlockerTwoThenDraw). A creature is blocking only from
// the declaration of blockers to the end of combat (CR 506.4), so the
// spell has no legal target outside that window. If the target is no
// longer legal on resolution the spell does nothing, the draw included
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3b2f1364-56e7-461a-b03f-0e14d83b0f5e",
		Name:         "Evasive Maneuvers",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target blocking creature you control", BlockingCreature(), YouControl()),
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2)),
		OnResolve:    pumpBlockerTwoThenDraw("Evasive Maneuvers — +2/+2"),
	})
}
