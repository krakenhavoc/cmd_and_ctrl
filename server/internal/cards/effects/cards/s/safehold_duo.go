package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Safehold Duo — Creature — Elf Warrior Shaman {3}{G/W} 2/4:
//
//	"Whenever you cast a green spell, this creature gets +1/+1 until end of turn.
//	 Whenever you cast a white spell, this creature gains vigilance until end of turn."
//
// Two triggers from the Duo cycle's shared helper (effects.S82OldB528Duo): a
// spell of both colours triggers both, and a spell of neither does nothing.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "76168d1b-266d-4d27-9e97-248b0c2c7e8f",
		Name:         "Safehold Duo",
		Completeness: effects.CompletenessFull,
		Triggered:    effects.S82OldB528Duo("Safehold Duo", "G", "W", "vigilance"),
	})
}
