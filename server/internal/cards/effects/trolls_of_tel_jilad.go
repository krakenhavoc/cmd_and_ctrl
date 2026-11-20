package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trolls of Tel-Jilad — Creature — Troll Shaman {5}{G}{G}, 5/6 (Mirrodin, #5606):
//
//	"{1}{G}: Regenerate target green creature."
//
// Yavimaya Hollow's regeneration of a target creature, narrowed to a
// green one (OfColor reads the creature's colours through the
// layers). The Trolls may target themselves. A target that is no
// longer green as the ability resolves gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4478e843-fd25-4ad1-93f8-2dee605386fc",
		Name:         "Trolls of Tel-Jilad",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}: Regenerate target green creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ManaCost("{1}{G}"),
			Targets: TargetCreature("target green creature", OfColor("G")),
			Effect:  regenerateTheTargetPermanent,
		}},
	})
}
