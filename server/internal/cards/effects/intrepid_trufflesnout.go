package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Intrepid Trufflesnout // Go Hog Wild — Creature — Boar {1}{G}, 3/1 //
// Instant — Adventure {1}{G} (Wilds of Eldraine, #4471):
//
//	Intrepid Trufflesnout — "Whenever this creature attacks alone, create
//	 a Food token. (It's an artifact with "{2}, {T}, Sacrifice this token:
//	 You gain 3 life.")"
//	Go Hog Wild — "Target creature gets +2/+2 until end of turn. (Then
//	 exile this card. You may cast the creature later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"). "Attacks alone" is the only creature declared as an attacker
// (CR 506.5, ThisAttackedAlone), counted over the whole declaration, so
// attacking beside any other creature makes no Food. The Food is the
// catalog's token (CR 111.10b). Go Hog Wild is the one-target pump until
// cleanup; the Adventure lifecycle is the engine's (CR 715.3d).
//
// No simplification.
const intrepidTrufflesnoutOracleID = "147d1e5c-fcfd-4115-a2f9-bea9b7ad7bb6"

func init() {
	Register(Spec{
		OracleID:     intrepidTrufflesnoutOracleID,
		Name:         "Intrepid Trufflesnout",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, ThisAttackedAlone, "Intrepid Trufflesnout — create a Food token",
				Do(CreateToken{Template: FoodToken(), N: 1})),
		},
	})
	Register(Spec{
		OracleID:     intrepidTrufflesnoutOracleID + "#1",
		Name:         "Go Hog Wild",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2)),
		Targets:      TargetCreature("target creature"),
		OnResolve:    boostTheTargetUntilEOT(2, 2, "Go Hog Wild — +2/+2"),
	})
}
