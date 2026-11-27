package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emberstrike Duo — Creature — Elemental Warrior Shaman {1}{B/R} 1/1:
//
//	"Whenever you cast a black spell, this creature gets +1/+1 until end of turn.
//	 Whenever you cast a red spell, this creature gains first strike until end of turn."
//
// Two triggers from the Duo cycle's shared helper (effects.S82OldB528Duo): a
// spell of both colours triggers both, and a spell of neither does nothing.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "44295009-9f85-4022-a685-a96079ba3e26",
		Name:         "Emberstrike Duo",
		Completeness: effects.CompletenessFull,
		Triggered:    effects.S82OldB528Duo("Emberstrike Duo", "B", "R", "first strike"),
	})
}
