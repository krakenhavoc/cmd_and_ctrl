package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "5a0eb270-b142-45da-87a2-2f1c4e25db17",
		Name:            "Chandra's Spitfire",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.AnOpponentWasDealtNoncombatDamage,
				"Chandra's Spitfire — +3/+0 until end of turn",
				effects.ThisGetsUntilEndOfTurn(3, 0, "Chandra's Spitfire — +3/+0 until end of turn")),
		},
	})
}
