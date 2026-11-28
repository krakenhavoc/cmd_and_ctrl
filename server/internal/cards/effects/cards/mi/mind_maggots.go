package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mind Maggots — Creature — Insect {3}{B}, 2/2 (Exodus, S82 OLD batch 741, #5626):
//
//	"When this creature enters, discard any number of creature cards. For
//	 each card discarded this way, put two +1/+1 counters on this creature."
//
// Borborygmos and Fblthp's discard prompt over the creature cards in hand
// (none in hand asks nothing); the counters are two for each card that
// really left the hand, put on the Maggots if it is still the same permanent
// as the tail runs.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fcaa5a54-eb73-4d78-96a3-826d1902f642",
		Name:         "Mind Maggots",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Mind Maggots — discard any number of creature cards; two +1/+1 counters for each", effects.S82OldB741MindMaggotsDiscard),
		},
	})
}
