package effects

// Faerie Guidemother // Gift of the Fae — Creature — Faerie {W}, 1/1 // Sorcery — Adventure {1}{W} (Throne of Eldraine, #5516):
//
//	Faerie Guidemother — "Flying"
//	Gift of the Fae — "Target creature gets +2/+1 and gains flying until end of turn."
//
// Both faces register: the creature under the oracle ID and the Adventure
// under "<oracle_id>#1". The Adventure lifecycle (exile on resolution, cast
// the creature later, CR 715) is the engine's.
//
// No simplification.
const faerieGuidemotherOracleID = "04fbb7aa-0dd4-4ac6-837d-822227fb6356"

func init() {
	Register(Spec{
		OracleID:        faerieGuidemotherOracleID,
		Name:            "Faerie Guidemother",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
	Register(Spec{
		OracleID:     faerieGuidemotherOracleID + "#1",
		Name:         "Gift of the Fae",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 1, "flying")),
		Targets:      TargetCreature("target creature"),
		OnResolve:    tspB1PumpAndGrantUntilEOT(2, 1, "flying", "Gift of the Fae"),
	})
}
