package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chandra's Spitfire — Creature — Elemental {2}{R}, 1/3 (Foundations
// Jumpstart):
//
//	"Flying
//	 Whenever an opponent is dealt noncombat damage, this creature gets
//	 +3/+0 until end of turn."
//
// The trigger is Massacre Girl, Most Wanted's: one per noncombat damage
// event to a player other than the Spitfire's controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5a0eb270-b142-45da-87a2-2f1c4e25db17",
		Name:            "Chandra's Spitfire",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, anOpponentWasDealtNoncombatDamage,
				"Chandra's Spitfire — +3/+0 until end of turn",
				thisGetsUntilEndOfTurn(3, 0, "Chandra's Spitfire — +3/+0 until end of turn")),
		},
	})
}
