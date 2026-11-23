package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deeproot Champion — Creature — Merfolk Shaman {1}{G}, 1/1 (Ixalan, #5532):
//
//	"Whenever you cast a noncreature spell, put a +1/+1 counter on this creature."
//
// Spellgorger Weird's trigger: casting a noncreature spell puts a
// +1/+1 counter on it, and the trigger resolves before the spell does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "507c2921-7070-482d-8754-dc667e3315bd",
		Name:         "Deeproot Champion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(), "Deeproot Champion — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
