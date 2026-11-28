package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wall of Souls — Creature — Wall {1}{B}, 0/4 (Stronghold, S82 OLD batch 747,
// #5627):
//
//	"Defender
//	 Whenever this creature is dealt combat damage, it deals that much damage
//	 to target opponent or planeswalker."
//
// Brash Taunter's trigger waiting for the combat damage to finish, so the TOTAL
// dealt together is "that much" (CR 120.3, CR 510.2), with Souls of the
// Faultless' combat-only reading and an opponent-or-planeswalker target chosen
// as the trigger goes on the stack. The Wall is the damage source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2c85c9ea-c3fd-4981-a65c-4405bd87764e",
		Name:            "Wall of Souls",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.ThisIsDealtDamageReflected("Wall of Souls — it deals that much damage to target opponent or planeswalker", true),
		},
	})
}
