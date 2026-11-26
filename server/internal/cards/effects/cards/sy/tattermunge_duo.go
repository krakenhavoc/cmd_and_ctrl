package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tattermunge Duo — Creature — Goblin Warrior Shaman {2}{R/G} 2/3:
//
//	"Whenever you cast a red spell, this creature gets +1/+1 until end of turn.
//	 Whenever you cast a green spell, this creature gains forestwalk until end of turn."
//
// Two triggers from the Duo cycle's shared helper (effects.S82OldB528Duo): a
// spell of both colours triggers both, and a spell of neither does nothing.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a56d13b5-a84d-4629-b9ec-4f407ee62596",
		Name:         "Tattermunge Duo",
		Completeness: effects.CompletenessFull,
		Triggered:    effects.S82OldB528Duo("Tattermunge Duo", "R", "G", "forestwalk"),
	})
}
