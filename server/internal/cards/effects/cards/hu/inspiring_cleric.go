package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inspiring Cleric — Creature — Vampire Cleric {2}{W}, 3/2 (Ixalan, #5532):
//
//	"When this creature enters, you gain 4 life."
//
// Spiritual Guardian's body: an enters trigger on the stack (CR 603.6a)
// that gains its controller 4 life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "65e19aec-e8ea-412f-a239-bd8d1b78249c",
		Name:         "Inspiring Cleric",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Inspiring Cleric — you gain 4 life", effects.Do(effects.GainLife{Amount: 4})),
		},
	})
}
