package effects

// Wielding the Green Dragon — Sorcery {1}{G} (Portal Three Kingdoms, #5622):
//
//	"Target creature gets +4/+4 until end of turn."
//
// Monstrous Growth's body (plusFourPlusFourToTheTarget). The pump ends at
// cleanup (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "22fa9020-783e-4292-857e-9d45a3458d71",
		Name:         "Wielding the Green Dragon",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 4, 4)),
		Targets:      TargetCreature("target creature"),
		OnResolve:    plusFourPlusFourToTheTarget("Wielding the Green Dragon"),
	})
}
