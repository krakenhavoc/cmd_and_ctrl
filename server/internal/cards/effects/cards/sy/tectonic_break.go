package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tectonic Break — Sorcery {X}{R}{R} (Mercadian Masques, S82 OLD batch 697, #5619):
//
//	"Each player sacrifices X lands of their choice."
//
// S82OldB697TectonicBreak: every living seat chooses X lands in APNAP order
// and they are sacrificed together; a player with fewer lands sacrifices
// what they have.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4a3c6d0e-6d77-4434-95f4-737879de026f",
		Name:         "Tectonic Break",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.S82OldB697TectonicBreak,
	})
}
