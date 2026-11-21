package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soul Feast — Sorcery {3}{B}{B} (Urza's Destiny, #5621):
//
//	"Target player loses 4 life and you gain 4 life."
//
// Syphon Life's drain at 4, through the shared
// targetPlayerLosesLifeYouGain body. It is life loss, not damage, so
// nothing prevents it. A target player who has left the game counters
// the spell (CR 608.2b), and you gain nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8186fd80-015f-470c-9e1c-cbf45764a057",
		Name:         "Soul Feast",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve:    effects.TargetPlayerLosesLifeYouGain(4),
	})
}
