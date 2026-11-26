package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thistledown Duo — Creature — Kithkin Soldier Wizard {2}{W/U} 2/2:
//
//	"Whenever you cast a white spell, this creature gets +1/+1 until end of turn.
//	 Whenever you cast a blue spell, this creature gains flying until end of turn."
//
// Two triggers from the Duo cycle's shared helper (effects.S82OldB528Duo): a
// spell of both colours triggers both, and a spell of neither does nothing.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f6777aee-c39b-49d8-9049-5d2b1e888dd0",
		Name:         "Thistledown Duo",
		Completeness: effects.CompletenessFull,
		Triggered:    effects.S82OldB528Duo("Thistledown Duo", "W", "U", "flying"),
	})
}
