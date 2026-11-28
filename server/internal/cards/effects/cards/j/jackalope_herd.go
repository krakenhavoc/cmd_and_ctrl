package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jackalope Herd — Creature — Rabbit Beast {3}{G}, 4/5 (Exodus, S82 OLD batch 744, #5626):
//
//	"When you cast a spell, return this creature to its owner's hand."
//
// A cast trigger on any spell you cast (Beast Whisperer's When, with no type
// filter) that returns the Herd if it is still the permanent that triggered
// (CR 400.7). It resolves before the spell it answers, and casting the Herd
// itself doesn't trigger it, since it isn't on the battlefield yet.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bcc8ce61-a881-42be-ab2c-64351d17f9f8",
		Name:         "Jackalope Herd",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(nil, "Jackalope Herd — return this creature to its owner's hand", effects.S82OldB744ReturnThisToHand),
		},
	})
}
