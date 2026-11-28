package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Evasive Action — Instant {1}{U} (Apocalypse, #5614):
//
//	"Domain — Counter target spell unless its controller pays {1} for each
//	 basic land type among lands you control."
//
// The count is the domain of the lands you control as it resolves (CR 305.6,
// read layered); the spell's controller is asked the CR 118.12 question.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4543a99d-eefa-470d-976d-11250524ae28",
		Name:         "Evasive Action",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve:    effects.S82OldB635EvasiveAction,
	})
}
