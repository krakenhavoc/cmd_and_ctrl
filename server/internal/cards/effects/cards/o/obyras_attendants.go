package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
const ObyrasAttendantsOracleID = "396b088d-f9af-4ee1-843f-dbe1633f9cc8"

func init() {
	effects.Register(effects.Spec{
		OracleID:        ObyrasAttendantsOracleID,
		Name:            "Obyra's Attendants",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
	effects.Register(effects.Spec{
		OracleID:     ObyrasAttendantsOracleID + "#1",
		Name:         "Desperate Parry",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -4, 0)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.BoostTheTargetUntilEOT(-4, 0, "Desperate Parry — -4/-0"),
	})
}
