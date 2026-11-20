package effects

// Garenbrig Carver // Shield's Might — Creature — Human Warrior {3}{G}, 3/2 // Instant — Adventure {1}{G} (Throne of Eldraine, #5516):
//
//	Garenbrig Carver — no rules text.
//	Shield's Might — "Target creature gets +2/+2 until end of turn."
//
// Both faces register: the creature under the oracle ID and the Adventure
// under "<oracle_id>#1". The Adventure lifecycle (exile on resolution, cast
// the creature later, CR 715) is the engine's.
//
// No simplification.
const garenbrigCarverOracleID = "496c9633-ff17-48df-9b24-1d8cbfd06126"

func init() {
	Register(Spec{
		OracleID:     garenbrigCarverOracleID,
		Name:         "Garenbrig Carver",
		Completeness: CompletenessFull,
	})
	Register(Spec{
		OracleID:     garenbrigCarverOracleID + "#1",
		Name:         "Shield's Might",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2)),
		Targets:      TargetCreature("target creature"),
		OnResolve:    boostTheTargetUntilEOT(2, 2, "Shield's Might — +2/+2 until end of turn"),
	})
}
