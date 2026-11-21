package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Midnight Guard — Creature — Human Soldier {2}{W}, 2/3 (Dark Ascension, #5571):
//
//	"Whenever another creature enters, untap this creature."
//
// Soul Warden's trigger condition (any other creature, under any player's
// control) with Fain's untap of the source (b31UntapSelf), which does
// nothing once the Guard has left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5430b0e-e77f-47b8-ae6d-30bcff4d7e35",
		Name:         "Midnight Guard",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AnotherCreatureEntered, "Midnight Guard — untap this creature", effects.B31UntapSelf),
		},
	})
}
