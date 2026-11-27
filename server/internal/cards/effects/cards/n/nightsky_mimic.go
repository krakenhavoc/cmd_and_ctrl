package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nightsky Mimic — Creature — Shapeshifter {1}{W/B} 2/1:
//
//	"Whenever you cast a spell that's both white and black, this creature has base
//	 power and toughness 4/4 until end of turn and gains flying until end of turn."
//
// One trigger (effects.S82OldB526Mimic, shared by the cycle): the spell's
// colours are read as it was cast, a spell of one of the colours alone does
// nothing, and the base P/T is a layer 7b change that counters and later
// pumps still add to.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed186aea-e0dc-4d88-a619-748b2ebf0c52",
		Name:         "Nightsky Mimic",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.S82OldB526Mimic("Nightsky Mimic", "W", "B", 4, 4, false, "flying"),
		},
	})
}
