package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sparkcaster — Creature — Kavu {2}{R}{G}, 5/3 (Planeshift, #5615):
//
//	"When this creature enters, return a red or green creature you
//	 control to its owner's hand.
//	 When this creature enters, it deals 1 damage to target player or
//	 planeswalker."
//
// Two enters triggers, which their controller orders as they go on the
// stack. The first is the shared gating return (gating.go), chosen as it
// resolves, the Kavu itself included. The second is Sunscorched Desert's
// ping with Boros Charm's "target player or planeswalker" clause, chosen
// as it goes on the stack (CR 603.3d). The Kavu deals that damage even
// after it has gone back to hand, from its last-known information.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "84c912e8-6a92-47f7-88b6-e777bd0bb424",
		Name:         "Sparkcaster",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			gatingReturn("Sparkcaster", "R", "G"),
			TriggerWithPurpose(Targeting(WhenThisEnters("Sparkcaster — 1 damage to target player or planeswalker",
				sourceDealsDamageToEachLegalTarget(1)), targetPlayerOrPlaneswalker()), ForTargets(DamageToTarget(0, 1))),
		},
	})
}
