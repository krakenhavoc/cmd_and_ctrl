package effects

// Obyra's Attendants // Desperate Parry — Creature — Faerie Wizard {4}{U},
// 3/4 // Instant — Adventure {1}{U} (Wilds of Eldraine, #4471):
//
//	Obyra's Attendants — "Flying"
//	Desperate Parry — "Target creature gets -4/-0 until end of turn. (Then
//	 exile this card. You may cast the creature later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"). Desperate Parry is Arcane Subtraction's shrink without the
// learn: power can go below zero in the layers and deals no damage there,
// toughness is untouched. The Adventure lifecycle is the engine's
// (CR 715.3d).
//
// No simplification.
const obyrasAttendantsOracleID = "396b088d-f9af-4ee1-843f-dbe1633f9cc8"

func init() {
	Register(Spec{
		OracleID:        obyrasAttendantsOracleID,
		Name:            "Obyra's Attendants",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
	Register(Spec{
		OracleID:     obyrasAttendantsOracleID + "#1",
		Name:         "Desperate Parry",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(HarmUntilEOT(0, -4, 0)),
		Targets:      TargetCreature("target creature"),
		OnResolve:    boostTheTargetUntilEOT(-4, 0, "Desperate Parry — -4/-0"),
	})
}
