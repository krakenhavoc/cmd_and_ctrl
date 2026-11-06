package effects

// Aang's Defense — Instant {W} (Avatar: The Last Airbender Eternal,
// #5075):
//
//	"Target blocking creature you control gets +2/+2 until end of turn.
//	 Draw a card."
//
// Captain's Defense with the target narrowed to a blocking creature you
// control. If the target is illegal on resolution the spell does nothing,
// so no card is drawn (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "92b617f9-6e6b-42a0-9c45-6ff5f6193051",
		Name:         "Aang's Defense",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target blocking creature you control", BlockingCreature(), YouControl()),
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2)),
		OnResolve:    pumpBlockerTwoThenDraw("Aang's Defense — +2/+2"),
	})
}
