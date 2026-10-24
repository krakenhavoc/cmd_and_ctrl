package effects

// Blessed Breath — Instant — Arcane {W}:
//
//	"Target creature you control gains protection from the color of your
//	choice until end of turn.
//	Splice onto Arcane {W}"
//
// Center Soul's body: the colour is chosen as the text resolves. Splice
// is ADR 0202 (#1989).
func init() {
	Register(Spec{
		OracleID:     "e83e1f99-1f9b-49fc-bdc7-5f6335e401e4",
		Name:         "Blessed Breath",
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "protection")),
		Completeness: CompletenessFull,
		Splice:       SpliceOntoArcane("{W}"),
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve:    protectionFromAChosenColorForTheTarget("Blessed Breath"),
	})
}
