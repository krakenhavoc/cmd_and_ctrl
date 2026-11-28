package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Word of Undoing — Instant {U} (Ice Age, S82 OLD, #5636):
//
//	"Return target creature and all white Auras you own attached to it to their owners' hands."
//
// Hubris's creature-and-Auras return, with the Auras limited to white ones the caster owns (not controls).
// All of them leave together.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2597a1ee-fe9e-44b5-a5e3-d34332c6c123",
		Name:         "Word of Undoing",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.S82OldB844WordOfUndoing,
	})
}
