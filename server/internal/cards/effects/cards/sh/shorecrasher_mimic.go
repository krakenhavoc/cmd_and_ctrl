package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shorecrasher Mimic — Creature — Shapeshifter {1}{G/U} 2/1:
//
//	"Whenever you cast a spell that's both green and blue, this creature has base
//	 power and toughness 5/3 until end of turn and gains trample until end of turn."
//
// One trigger (effects.S82OldB526Mimic, shared by the cycle): the spell's
// colours are read as it was cast, a spell of one of the colours alone does
// nothing, and the base P/T is a layer 7b change that counters and later
// pumps still add to.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3da837db-b3e9-46e4-939d-54ab43af94d9",
		Name:         "Shorecrasher Mimic",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.S82OldB526Mimic("Shorecrasher Mimic", "G", "U", 5, 3, false, "trample"),
		},
	})
}
