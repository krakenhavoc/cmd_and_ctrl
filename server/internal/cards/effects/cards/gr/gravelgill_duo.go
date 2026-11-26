package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gravelgill Duo — Creature — Merfolk Rogue Warrior {2}{U/B} 2/1:
//
//	"Whenever you cast a blue spell, this creature gets +1/+1 until end of turn.
//	 Whenever you cast a black spell, this creature gains fear until end of turn."
//
// Two triggers from the Duo cycle's shared helper (effects.S82OldB528Duo): a
// spell of both colours triggers both, and a spell of neither does nothing.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a9e28134-382b-45e5-8fd0-cc0e7d56fd0c",
		Name:         "Gravelgill Duo",
		Completeness: effects.CompletenessFull,
		Triggered:    effects.S82OldB528Duo("Gravelgill Duo", "U", "B", "fear"),
	})
}
