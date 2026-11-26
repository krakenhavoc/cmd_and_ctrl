package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alpha Brawl — Sorcery {6}{R}{R}:
//
//	"Target creature an opponent controls deals damage equal to its power to
//	 each other creature that player controls, then each of those creatures
//	 deals damage equal to its power to that creature."
//
// Both halves are simultaneous damage events; the creatures are fixed as the
// spell resolves, so a creature the first half kills still hits back (state-based
// actions wait until the spell is done, CR 704.3).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cfc023e0-9e9d-48ac-bfb5-74bfd5977dee",
		Name:         "Alpha Brawl",
		Completeness: effects.CompletenessFull,
		Targets:      effects.Harmful(effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		OnResolve:    effects.OldB442AlphaBrawl,
	})
}
