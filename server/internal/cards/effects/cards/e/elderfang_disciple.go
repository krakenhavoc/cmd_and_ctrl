package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elderfang Disciple — Creature — Elf Cleric {1}{B}, 1/1 (Foundations Jumpstart):
//
//	"When this creature enters, each opponent discards a card."
//
// Burglar Rat's body: one discard prompt per opponent, each offering only
// their own hand; an opponent with no cards is skipped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8db63f1a-0d63-4b11-966d-818d83a082f9",
		Name:         "Elderfang Disciple",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Elderfang Disciple — each opponent discards a card", effects.EachOpponentDiscardsOne),
		},
	})
}
