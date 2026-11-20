package effects

// Stave Off — Instant {W} (Magic 2012, #5573):
//
//	"Target creature gains protection from the color of your choice
//	 until end of turn."
//
// Blessed Breath's body (protectionFromAChosenColorForTheTarget) on any
// target creature: the color is chosen as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c77c23bc-efde-43ce-bf6f-7a8531d81d89",
		Name:         "Stave Off",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "protection")),
		Targets:      TargetCreature("target creature"),
		OnResolve:    protectionFromAChosenColorForTheTarget("Stave Off"),
	})
}
