package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Parallax Inhibitor — Artifact {2} (Nemesis, S82 OLD batch 680, #5618):
//
//	"{1}, {T}, Sacrifice this artifact: Put a fade counter on each permanent
//	 with fading you control."
//
// "Permanent with fading" is one whose catalog row carries the fading upkeep
// trigger (CR 702.32a). The Inhibitor is sacrificed as a cost, so it is
// itself no longer there to count; the permanents are chosen as the ability
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b8944199-9d33-4126-b529-b52e9f0116ff",
		Name:         "Parallax Inhibitor",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice this artifact: Put a fade counter on each permanent with fading you control.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeThis()),
			Effect:  effects.OldB680FadeCounterOnEachFadingPermanentYouControl,
		}},
	})
}
