package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cinder Hellion — Creature — Hellion {4}{R}, 4/4 (Oath of the Gatewatch, #5545):
//
//	"Trample
//	 When this creature enters, it deals 2 damage to target opponent or
//	 planeswalker."
//
// The enters trigger targets as it goes on the stack (CR 603.3d), with
// Zealot of the God-Pharaoh's "target opponent or planeswalker" clause,
// and the damage is dealt by the Hellion.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9e1ad977-2814-4336-93ec-57b7736a629f",
		Name:            "Cinder Hellion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Cinder Hellion — 2 damage to target opponent or planeswalker", DealDamageToTheTarget(2)),
				targetOpponentOrPlaneswalker()),
		},
	})
}
