package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Burst of Energy — Instant {W}:
//
//	"Untap target permanent."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b795ecfd-31ac-4a2a-9679-97b088932d93",
		Name:         "Burst of Energy",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target permanent"),
		OnResolve:    effects.OldB718UntapTheTarget,
	})
}
