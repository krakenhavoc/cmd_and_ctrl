package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// War Report — Instant {3}{W}:
//
//	"You gain life equal to the number of creatures on the battlefield plus
//	 the number of artifacts on the battlefield."
//
// Every player's permanents count, and an artifact creature counts once as a
// creature and once as an artifact (read through the layered view).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "45dfb358-cd16-468e-bc89-ff17c6e244c1",
		Name:         "War Report",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.WarReportGain,
	})
}
