package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "a3bf5d18-77d5-4881-916c-cc125e621b6e",
		Name:             "Aphetto Exterminator",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Morph("{3}{B}")},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(
				WhenThisIsTurnedFaceUp("Aphetto Exterminator — target creature gets -3/-3 until end of turn",
					boostTargetUntilEOT("Aphetto Exterminator — -3/-3 until end of turn", -3, -3)),
				TargetCreature("target creature")),
				ForTargets(HarmUntilEOT(0, -3, -3))),
		},
	})
}
