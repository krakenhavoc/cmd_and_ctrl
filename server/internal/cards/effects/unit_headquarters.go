package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// UNIT Headquarters — Plane — Earth (Doctor Who, #4870):
//
//	"When you planeswalk to UNIT Headquarters and at the beginning of your
//	 upkeep, create a tapped 1/1 white Soldier creature token.
//	 Whenever chaos ensues, put a +1/+1 counter on each creature you
//	 control."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6).
// The Soldier enters tapped. The chaos ability (CR 311.7) puts one counter
// on each creature the planar controller controls as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a7641877-1a61-467c-9365-487cf5a5656d",
		Name:         "UNIT Headquarters",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThisAndAtYourUpkeep("UNIT Headquarters — create a tapped 1/1 white Soldier creature token",
				Do(CreateTokenAdvanced{Spec: TokenSpec{Template: TokenCard("1/1 white Soldier"), Tapped: true}, N: 1})),
			WheneverChaosEnsues("UNIT Headquarters — put a +1/+1 counter on each creature you control", b11PutCounterOnEachCreatureYouControl),
		},
	})
}
