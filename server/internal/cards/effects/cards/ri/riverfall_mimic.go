package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Riverfall Mimic — Creature — Shapeshifter {1}{U/R} 2/1:
//
//	"Whenever you cast a spell that's both blue and red, this creature has base
//	 power and toughness 3/3 until end of turn and can't be blocked this turn."
//
// One trigger (effects.S82OldB526Mimic, shared by the cycle): the spell's
// colours are read as it was cast, a spell of one of the colours alone does
// nothing, and the base P/T is a layer 7b change that counters and later
// pumps still add to.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "16948dd9-959b-4972-b337-37ecc2b141bd",
		Name:         "Riverfall Mimic",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.S82OldB526Mimic("Riverfall Mimic", "U", "R", 3, 3, true, ""),
		},
	})
}
