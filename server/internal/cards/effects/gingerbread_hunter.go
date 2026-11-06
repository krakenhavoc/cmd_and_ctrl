package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gingerbread Hunter // Puny Snack — Creature — Giant {4}{G}, 5/5 //
// Instant — Adventure {2}{B} (Wilds of Eldraine, #4471):
//
//	Gingerbread Hunter — "When this creature enters, create a Food token.
//	 (It's an artifact with "{2}, {T}, Sacrifice this token: You gain 3
//	 life.")"
//	Puny Snack — "Target creature gets -2/-2 until end of turn. (Then
//	 exile this card. You may cast the creature later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"). The entry trigger makes the catalog's Food token
// (CR 111.10b); Puny Snack is the one-target shrink until cleanup. The
// Adventure lifecycle is the engine's (CR 715.3d).
//
// No simplification.
const gingerbreadHunterOracleID = "f762c225-1c8c-4f12-aa88-c8c89e0a7185"

func init() {
	Register(Spec{
		OracleID:     gingerbreadHunterOracleID,
		Name:         "Gingerbread Hunter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Gingerbread Hunter — create a Food token", Do(CreateToken{Template: FoodToken(), N: 1})),
		},
	})
	Register(Spec{
		OracleID:     gingerbreadHunterOracleID + "#1",
		Name:         "Puny Snack",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(HarmUntilEOT(0, -2, -2)),
		Targets:      TargetCreature("target creature"),
		OnResolve:    boostTheTargetUntilEOT(-2, -2, "Puny Snack — -2/-2"),
	})
}
