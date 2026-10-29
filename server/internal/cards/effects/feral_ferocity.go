package effects

// Feral Ferocity — Instant {1}{G} (Marvel Super Heroes, #4455):
//
//	"Target creature gets +4/+4 until end of turn."
//
// Titanic Growth under a Marvel name: the target clause is plain "target
// creature", so an opponent's creature is a legal target, and the boost
// is plusFourPlusFourToTheTarget.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5e968bf3-8d83-480b-81a4-98883cafd04f",
		Name:         "Feral Ferocity",
		Purpose:      ForTargets(BenefitUntilEOT(0, 4, 4)),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve:    plusFourPlusFourToTheTarget("Feral Ferocity"),
	})
}
