package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fungusaur — Creature — Fungus Dinosaur {3}{G} 2/2 (Eighth Edition, S82 older sets batch 620, #5607):
//
//	"Whenever this creature is dealt damage, put a +1/+1 counter on it."
//
// Snapping Sailback's enrage trigger. One trigger per batch of damage dealt
// together, as Snapping Sailback's is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3e771f5b-2de3-4a1b-8281-ab1f7491e5a1",
		Name:         "Fungusaur",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Cmdr2B08EnrageTrigger("Fungusaur — put a +1/+1 counter on it", effects.PlusOneCountersOnThis(1)),
		},
	})
}
