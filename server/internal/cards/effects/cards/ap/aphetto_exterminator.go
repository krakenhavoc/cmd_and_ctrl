package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aphetto Exterminator — Creature — Human Wizard {2}{B}, 3/1 (Legions,
// #5609):
//
//	"Morph {3}{B}
//	 When this creature is turned face up, target creature gets -3/-3
//	 until end of turn."
//
// The morph trigger (CR 708.8) with Ambush Gigapede's shrink on its
// target. Face down it has no text (CR 708.2a), so casting it face down
// does nothing; the trigger comes with the turn face up. A target gone
// by resolution is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "a3bf5d18-77d5-4881-916c-cc125e621b6e",
		Name:             "Aphetto Exterminator",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{3}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.WhenThisIsTurnedFaceUp("Aphetto Exterminator — target creature gets -3/-3 until end of turn",
					effects.BoostTargetUntilEOT("Aphetto Exterminator — -3/-3 until end of turn", -3, -3)),
				effects.TargetCreature("target creature")),
				effects.ForTargets(effects.HarmUntilEOT(0, -3, -3))),
		},
	})
}
