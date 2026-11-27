package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jaded Response — Instant {1}{U} (Apocalypse, #5614):
//
//	"Counter target spell if it shares a color with a creature you control."
//
// Any spell may be targeted; whether it shares a colour with a creature you
// control is judged as this resolves (CR 608.2b), so a spell that is colourless
// or shares none is left alone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "23917517-7561-45b3-b156-77cd2f340466",
		Name:         "Jaded Response",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve:    effects.S82OldB635JadedResponse,
	})
}
